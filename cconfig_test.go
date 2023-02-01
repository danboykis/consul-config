package consulconfig

import (
	"fmt"
	"strings"
	"testing"
)

type Foo struct {
	Url     string            `config:"url" json:"url"`
	Headers map[string]string `config:"headers" json:"headers"`
}
type Auth struct {
	Username string `config:"username"`
	Password string `config:"password"`
}
type Config struct {
	Foo      Foo      `config:"foo" json:"foo"`
	Service1 Auth     `config:"service1"`
	Service2 Auth     `config:"service2"`
	LogFile  string   `config:"logfile,/var/app/log/logfile.log" json:"logFile"`
	Refresh  bool     `config:"refresh" json:"refresh"`
	Items    []string `config:"items" json:"items"`
}

func TestConfigTree(t *testing.T) {
	m := map[string]string{
		"service1/username": "service1_username",
		"service1/password": "service1_password",
		"service2/username": "service2_username",
		"service2/password": "service2_password",
	}
	ct := newConfigTreeFromMap(m)

	paths := [][]string{{"service1", "username"}, {"service2", "username"}, {"service1", "password"}, {"service2", "password"}}

	for _, p := range paths {
		if v, e := ct.getInString(p...); !e || v != strings.Join(p, "_") {
			t.Errorf("wrong %s: %s", p[1], v)
		}
	}
}

func TestPopulateConfig(t *testing.T) {
	m := map[string]string{
		"foo/url":           "http://example.com/consul",
		"foo/headers":       "{\"foo\": \"bar\", \"baz\": \"quux\"}",
		"service1/username": "service1_username",
		"service1/password": "service1_password",
		"service2/username": "service2_username",
		"service2/password": "service2_password",
		"log-file":          "/var/app/log/logfile.log",
		"refresh":           "true",
		"items":             "[\"a\",\"b\"]",
	}
	ct := newConfigTreeFromMap(m)
	conf := &Config{}
	PopulateConfig(ct, conf)
	fmt.Printf("%+v\n", *conf)

	if conf.Service1.Username != "service1_username" {
		t.Errorf("service1 username is %v\n", "service1_username")
	}
	if conf.Service1.Password != "service1_password" {
		t.Errorf("service1 password is %v\n", "service1_password")
	}
	if conf.Service2.Username != "service2_username" {
		t.Errorf("service1 username is %v\n", "service2_username")
	}
	if conf.Service2.Password != "service2_password" {
		t.Errorf("service1 password is %v\n", "service2_password")
	}
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
	if len(conf.Items) != 2 || conf.Items[0] != "a" || conf.Items[1] != "b" {
		t.Errorf("did not parse items correctly: %v", conf.Items)
	}
}
