package browser

import (
	"bytes"
	"context"
	"errors"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/disintegration/imaging"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"image"
	"image/jpeg"
	"io/ioutil"
	"math"
	"os"
)

type Browser struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	TaskCtx     context.Context
	taskCancel  context.CancelFunc
	browser     *chromedp.Browser
	TempDir     string
	opts        []chromedp.ExecAllocatorOption
	log         *logging.Logger
}

func NewBrowser(execOptions map[string]interface{}, log *logging.Logger) (*Browser, error) {
	browser := &Browser{log: log}
	return browser, browser.Init(execOptions)
}

func (browser *Browser) Startup() error {
	// create the execution context
	browser.allocCtx, browser.allocCancel = chromedp.NewExecAllocator(context.Background(), browser.opts...)

	// also set up a custom logger
	browser.TaskCtx, browser.taskCancel = chromedp.NewContext(browser.allocCtx, chromedp.WithLogf(browser.log.Debugf))

	return nil
}

func (browser *Browser) Init(execOptions map[string]interface{}) error {
	var err error

	// create temporary directory
	browser.TempDir, err = ioutil.TempDir("", "bremote")
	if err != nil {
		return emperror.Wrap(err, "cannot create tempdir")
	}

	// build browser options
	browser.opts = append(chromedp.DefaultExecAllocatorOptions[:], chromedp.UserDataDir(browser.TempDir))
	for name, value := range execOptions {
		browser.opts = append(browser.opts, chromedp.Flag(name, value))
	}

	return browser.Startup()
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Liberally copied from puppeteer's source.
//
// Note: this will override the viewport emulation settings.
func fullScreenshot(quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}

func (browser *Browser) Screenshot(width int, height int, sigma float64) ([]byte, string, error){
	if !browser.IsRunning() {
		return nil, "", errors.New("browser not running")
	}
	var bufIn []byte
	if err := chromedp.Run(browser.TaskCtx, fullScreenshot(90, &bufIn)); err != nil {
		return nil, "", emperror.Wrapf(err, "cannot take screenshot")
	}
	// full size - no action
	if width == 0 && height == 0 {
		return bufIn, "image/png", nil
	}
	rawImage, _, err := image.Decode(bytes.NewReader(bufIn))
	if err != nil {
		return nil, "", emperror.Wrapf(err, "cannot decode png")
	}
	newraw := imaging.Resize(rawImage, width, 0, imaging.Lanczos)
	var sharpraw image.Image
	if sigma > 0.0 {
		sharpraw = imaging.Sharpen(newraw, sigma)
	} else {
		sharpraw = newraw
	}
	bufOut := []byte{}
	bwriter := bytes.NewBuffer(bufOut)
	if err := jpeg.Encode(bwriter, sharpraw, nil); err != nil {
		return nil, "", emperror.Wrapf(err, "cannot encode jpeg")
	}
	bufOut = bwriter.Bytes()
	return bufOut, "image/jpeg", nil
}

func (browser *Browser) IsRunning() bool {
	if browser.TaskCtx == nil {
		return false
	}
	return browser.TaskCtx.Err() == nil
}

func (browser *Browser) Tasks(tasks chromedp.Tasks) error {
	if err := browser.TaskCtx.Err(); err != nil {
		if err := browser.Startup(); err != nil {
			return emperror.Wrap(err, "cannot re-initialize browser")
		}
		if err := browser.Run(); err != nil {
			return emperror.Wrap(err, "cannot re-start browser")
		}
	}
	return chromedp.Run(browser.TaskCtx, tasks)
}

func (browser *Browser) Run() error {
	browser.log.Debug("running browser")
	if err := chromedp.Run(browser.TaskCtx); err != nil {
		return emperror.Wrap(err, "cannot start chrome")
	}
	return nil
}

func (browser *Browser) Close() {
	// all paranoia...
	browser.log.Debug("closing browser")

	if browser.taskCancel != nil {
		browser.taskCancel()
	}
	if browser.allocCancel != nil {
		browser.allocCancel()
	}
	if browser.TempDir != "" {
		os.RemoveAll(browser.TempDir)
	}
}
