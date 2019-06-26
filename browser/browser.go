package browser

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
)

type Browser struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	taskCtx     context.Context
	taskCancel  context.CancelFunc
	TempDir     string
	opts        []chromedp.ExecAllocatorOption
	log         *logging.Logger
}

func NewBrowser(execOptions map[string]interface{}, log *logging.Logger) (*Browser, error) {
	browser := &Browser{log: log}
	return browser, browser.Init(execOptions)
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

	// create the execution context
	browser.allocCtx, browser.allocCancel = chromedp.NewExecAllocator(context.Background(), browser.opts...)

	// also set up a custom logger
	browser.taskCtx, browser.taskCancel = chromedp.NewContext(browser.allocCtx, chromedp.WithLogf(browser.log.Debugf))

	return nil
}

func (browser *Browser) Run() error {
	browser.log.Debug("running browser")
	if err := chromedp.Run(browser.taskCtx); err != nil {
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
