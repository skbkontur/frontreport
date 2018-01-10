package sourcemap

import (
	"fmt"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/go-sourcemap/sourcemap"
	"github.com/patrickmn/go-cache"
	"github.com/skbkontur/frontreport"
)

// Processor converts stacktrace to readable format using sourcemaps
type Processor struct {
	Providers     string
	Logger        frontreport.Logger
	cache         *cache.Cache
	smapURLRegexp *regexp.Regexp
}

// Start initializes sourcemaps cache
func (p *Processor) Start() error {
	p.cache = cache.New(24*time.Hour, time.Hour)
	p.smapURLRegexp = regexp.MustCompile(`sourceMappingURL=(\S+)\s+$`)
	return nil
}

// Stop does nothing
func (p *Processor) Stop() error {
	return nil
}

// ProcessStack converts stacktrace frames to readable format
func (p *Processor) ProcessStack(stack []frontreport.StacktraceJSStackframe) []frontreport.StacktraceJSStackframe {
	processedStack := make([]frontreport.StacktraceJSStackframe, len(stack))
	for i := range stack {
		var sMap *sourcemap.Consumer

		cachedMap, found := p.cache.Get(stack[i].FileName)
		if found {
			sMap = cachedMap.(*sourcemap.Consumer)
		} else {
			var err error
			sMap, err = p.getMapFromJSURL(stack[i].FileName)
			if err != nil {
				p.Logger.Log("msg", "failed to get sourcemap from url", "error", err, "url", stack[i].FileName)
				processedStack[i] = stack[i]
				continue
			}
			p.cache.SetDefault(stack[i].FileName, sMap)
		}

		file, fn, line, col, ok := sMap.Source(stack[i].LineNumber, stack[i].ColumnNumber)
		if ok {
			processedStack[i] = frontreport.StacktraceJSStackframe{
				FileName:     file,
				FunctionName: fn,
				LineNumber:   line,
				ColumnNumber: col,
			}
			if processedStack[i].FunctionName == "" {
				processedStack[i].FunctionName = stack[i].FunctionName
			}
		} else {
			processedStack[i] = stack[i]
		}
	}
	return processedStack
}

func (p *Processor) getMapFromJSURL(jsURL string) (*sourcemap.Consumer, error) {
	if err := p.checkIfTrusted(jsURL); err != nil {
		return nil, err
	}

	jsResp, err := http.Get(jsURL)
	if err != nil {
		return nil, err
	}
	defer jsResp.Body.Close()

	jsBody, err := ioutil.ReadAll(jsResp.Body)
	if err != nil {
		return nil, err
	}

	matches := p.smapURLRegexp.FindSubmatch(jsBody)
	if len(matches) < 2 {
		return nil, errors.New("failed to find sourcemap URL in JS file")
	}
	smapPartialURL := string(matches[1])

	baseURL, err := url.Parse(jsURL)
	if err != nil {
		return nil, err
	}

	smapURL, err := baseURL.Parse(smapPartialURL)
	if err != nil {
		return nil, err
	}

	smapUrlString := smapURL.String()
	if err := p.checkIfTrusted(smapUrlString); err != nil {
		return nil, err
	}

	smapResp, err := http.Get(smapUrlString)
	if err != nil {
		return nil, err
	}
	defer smapResp.Body.Close()

	smapBody, err := ioutil.ReadAll(smapResp.Body)
	if err != nil {
		return nil, err
	}

	return sourcemap.Parse(smapURL.String(), smapBody)
}

func (p *Processor) checkIfTrusted(urlToCheck string) error {
	if matched, _ := regexp.MatchString(p.Providers, urlToCheck); matched {
		return nil
	}
	return fmt.Errorf("%s doesn't match trusted providers pattern: %s", urlToCheck, p.Providers)
}
