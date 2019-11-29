package browser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"reflect"
	"time"

	//"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/disintegration/imaging"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"golang.org/x/sync/semaphore"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
)

type Browser struct {
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	TaskCtx       context.Context
	taskCancel    context.CancelFunc
	browser       *chromedp.Browser
	TempDir       string
	opts          []chromedp.ExecAllocatorOption
	log           *logging.Logger
	semScreenshot *semaphore.Weighted
	browserLog    func(string, ...interface{})
}

// MouseAction are mouse input event actions
type MouseAction chromedp.Action

func NewBrowser(execOptions map[string]interface{}, log *logging.Logger, browserLogFunc func(string, ...interface{})) (*Browser, error) {
	browser := &Browser{
		log:           log,
		semScreenshot: semaphore.NewWeighted(1),
		browserLog:    browserLogFunc,
	}
	return browser, browser.Init(execOptions)
}

func (browser *Browser) getTimeoutCtx(duration time.Duration) context.Context {
	newCtx, _ := context.WithTimeout(browser.TaskCtx, 10*time.Second)
	return newCtx
}

func (browser *Browser) Startup() error {
	// create the execution context
	browser.allocCtx, browser.allocCancel = chromedp.NewExecAllocator(context.Background(), browser.opts...)

	// also set up a custom logger
	browser.TaskCtx, browser.taskCancel = chromedp.NewContext(browser.allocCtx, chromedp.WithLogf(browser.log.Debugf))

	chromedp.ListenTarget(browser.TaskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			str := fmt.Sprintf("%s - %s: ", ev.Timestamp.Time().Format(`2006-01-02T15:04:05`), ev.Type)
			for idx, arg := range ev.Args {
				if idx > 0 {
					str += ` // `
				}
				str += fmt.Sprintf("[%s]%s", arg.Type, arg.Value)
			}
			browser.browserLog(str)
		case *target.EventTargetDestroyed:
		case *cdproto.Message:
		case *target.EventTargetInfoChanged:
		case *target.EventTargetCreated:
		case *runtime.EventExecutionContextDestroyed:
		case *runtime.EventExecutionContextsCleared:
		case *runtime.EventExecutionContextCreated:
		case *dom.EventDocumentUpdated:
		case *dom.EventChildNodeInserted:
		case *dom.EventChildNodeCountUpdated:
		case *css.EventStyleSheetAdded:
		case *css.EventMediaQueryResultChanged:
		case *css.EventStyleSheetRemoved:
		case *page.EventFrameStoppedLoading:
		case *page.EventLoadEventFired:
		case *page.EventDomContentEventFired:
		case *page.EventFrameNavigated:
		case *page.EventFrameStartedLoading:
		case *log.EventEntryAdded:
		default:
			browser.browserLog(fmt.Sprintf("Event of type %v happened", reflect.TypeOf(ev)))
		}
	})
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
func fullScreenshot(quality int64, res *[]byte, logger *logging.Logger) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			layoutViewport, visualViewport, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}
			logger.Debugf("layoutViewport: %v", layoutViewport)
			logger.Debugf("visualViewport: %v", visualViewport)
			logger.Debugf("contentSize: %v", contentSize)
			// capture screenshot
			*res, err = page.CaptureScreenshot().Do(ctx)
			if err != nil {
				return emperror.Wrapf(err, "cannot capture screenshot")
			}
			return nil
		}),
	}
}

func (browser *Browser) Screenshot(width int, height int, sigma float64) ([]byte, string, error) {
	if !browser.IsRunning() {
		return nil, "", errors.New("browser not running")
	}
	// screenshot is resource intense. disallow parallel use
	if !browser.semScreenshot.TryAcquire(1) {
		return nil, "", errors.New("cannot acquire semaphore")
	}
	browser.log.Debugf("acquire semaphore")
	defer func() {
		browser.semScreenshot.Release(1)
		browser.log.Debugf("release semaphore")
	}()

	var bufIn []byte
	if err := chromedp.Run(browser.getTimeoutCtx(30*time.Second), fullScreenshot(90, &bufIn, browser.log)); err != nil {
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

// checks whether browser is running. if not, clean up
func (browser *Browser) IsRunning() bool {
	if browser.TaskCtx == nil {
		return false
	}
	if browser.TaskCtx.Err() != nil {
		browser.Close()
		return false
	}
	return true
}

func (browser *Browser) Tasks(tasks chromedp.Tasks) error {
	// screenshot is resource intense. wait until done...
	browser.semScreenshot.Acquire(context.Background(), 1)
	browser.log.Debugf("acquire semaphore")
	defer func() {
		browser.semScreenshot.Release(1)
		browser.log.Debugf("release semaphore")
	}()

	if !browser.IsRunning() {
		if err := browser.Startup(); err != nil {
			return emperror.Wrap(err, "cannot re-initialize browser")
		}
		if err := browser.Run(); err != nil {
			return emperror.Wrap(err, "cannot re-start browser")
		}
	}
	// run the task in background and return after task is done or timeoiut
	c1 := make(chan bool, 1)
	go func() {
		browser.log.Debugf("tasks started")
		if err := chromedp.Run(browser.TaskCtx, tasks); err != nil {
			browser.log.Errorf("error running task: %v", err)
			c1 <- false
			return
		}
		c1 <- true
	}()
	select {
	case res := <-c1:
		browser.log.Debugf("tasks returned: %v", res)
	case <-time.After(5 * time.Second):
		browser.log.Debugf("tasks timed out")
	}
	return nil
}

func (browser *Browser) Run() error {
	browser.log.Debug("running browser")
	c1 := make(chan bool, 1)
	go func() {
		browser.log.Debugf("tasks started")
		if err := chromedp.Run(browser.TaskCtx); err != nil {
			browser.log.Errorf("cannot start chrome: %v", err)
			c1 <- false
			return
		}
		c1 <- true
	}()
	select {
	case res := <-c1:
		browser.log.Debugf("tasks returned: %v", res)
	case <-time.After(5 * time.Second):
		browser.log.Debugf("tasks timed out")
	}
	return nil
}

func (browser *Browser) Close() {
	// all paranoia...
	browser.log.Debug("closing browser")

	if browser.taskCancel != nil {
		browser.taskCancel()
		browser.taskCancel = nil
	}
	if browser.allocCancel != nil {
		browser.allocCancel()
		browser.allocCancel = nil
	}
	if browser.TempDir != "" {
		os.RemoveAll(browser.TempDir)
	}
	browser.TaskCtx = nil
}

func (browser *Browser) MouseClickXY(x, y int64) error {
	// screenshot is resource intense. wait until done...
	browser.semScreenshot.Acquire(context.Background(), 1)
	browser.log.Debugf("acquire semaphore")
	defer func() {
		browser.semScreenshot.Release(1)
		browser.log.Debugf("release semaphore")
	}()

	if !browser.IsRunning() {
		if err := browser.Startup(); err != nil {
			return emperror.Wrap(err, "cannot re-initialize browser")
		}
		if err := browser.Run(); err != nil {
			return emperror.Wrap(err, "cannot re-start browser")
		}
	}
	// run the task in background and return after task is done or timeoiut
	c1 := make(chan bool, 1)
	go func() {
		browser.log.Debugf("mouseclick started")
		if err := chromedp.Run(browser.TaskCtx, MouseClickXYAction(float64(x), float64(y))); err != nil {
			browser.log.Errorf("error running mouseclick: %v", err)
			c1 <- false
			return
		}
		c1 <- true
	}()
	select {
	case res := <-c1:
		browser.log.Debugf("mousclick returned: %v", res)
	case <-time.After(5 * time.Second):
		browser.log.Debugf("mouseclick timed out")
	}
	return nil
}

// MouseClickXYAction is an action that sends a left mouse button click (ie,
// mousePressed and mouseReleased event) to the X, Y location.
func MouseClickXYAction(x, y float64, opts ...chromedp.MouseOption) MouseAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		p := &input.DispatchMouseEventParams{
			Type:       input.MousePressed,
			X:          x,
			Y:          y,
			Button:     input.ButtonLeft,
			ClickCount: 1,
		}

		// apply opts
		for _, o := range opts {
			p = o(p)
		}

		if err := p.Do(ctx); err != nil {
			return err
		}

		p.Type = input.MouseReleased
		return p.Do(ctx)
	})
}