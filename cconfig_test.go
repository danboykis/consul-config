package consulconfig

import (
	"fmt"
	"testing"
)

type Foo struct {
	Url     string            `config:"foo/url" json:"url"`
	Headers map[string]string `config:"foo/headers" json:"headers"`
}
type Config struct {
	Foo     Foo    `config:"foo" json:"foo"`
	LogFile string `config:"logfile,/var/app/log/logfile.log" json:"logFile"`
	Refresh bool   `config:"refresh" json:"refresh"`
}

func TestPopulateConfig(t *testing.T) {
	m := map[string]string{
		"foo/url":     "http://example.com/consul",
		"foo/headers": "{\"foo\": \"bar\", \"baz\": \"quux\"}",
		"log-file":    "/var/app/log/logfile.log",
		"refresh":     "true",
	}
	conf := &Config{}
	PopulateConfig(m, conf)
	fmt.Printf("%+v\n", *conf)
	
	if !conf.Refresh {
		t.Errorf("refresh is %v\n", conf.Refresh)
	}
	if conf.Foo.Url != m["foo/url"] {
		t.Errorf("Url is %v\n", conf.Foo.Url)
	}
	if v, exists := conf.Foo.Headers["foo"]; !exists || v != "bar" {
		t.Errorf("foo key is wrong %v\n", v)
	}
	if v, exists := conf.Foo.Headers["baz"]; !exists || v != "quux" {
		t.Errorf("baz key is wrong %v\n", v)
	}
}
