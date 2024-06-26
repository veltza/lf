package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type iconMap struct {
	icons         map[string]string
	useLinkTarget bool
}

func parseIcons() iconMap {
	im := iconMap{
		icons:         make(map[string]string),
		useLinkTarget: false,
	}

	defaultIcons := []string{
		"ln=l",
		"or=l",
		"tw=t",
		"ow=d",
		"st=t",
		"di=d",
		"pi=p",
		"so=s",
		"bd=b",
		"cd=c",
		"su=u",
		"sg=g",
		"ex=x",
		"fi=-",
	}

	im.parseEnv(strings.Join(defaultIcons, ":"))

	if env := os.Getenv("LF_ICONS"); env != "" {
		im.parseEnv(env)
	}

	for _, path := range gIconsPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			im.parseFile(path)
		}
	}

	return im
}

func (im *iconMap) parseFile(path string) {
	log.Printf("reading file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		log.Printf("opening icons file: %s", err)
		return
	}
	defer f.Close()

	pairs, err := readPairs(f)
	if err != nil {
		log.Printf("reading icons file: %s", err)
		return
	}

	for _, pair := range pairs {
		im.parsePair(pair)
	}
}

func (im *iconMap) parseEnv(env string) {
	for _, entry := range strings.Split(env, ":") {
		if entry == "" {
			continue
		}

		pair := strings.Split(entry, "=")

		if len(pair) != 2 {
			log.Printf("invalid $LF_ICONS entry: %s", entry)
			return
		}

		im.parsePair(pair)
	}
}

func (im *iconMap) parsePair(pair []string) {
	key, val := pair[0], pair[1]

	key = replaceTilde(key)

	if filepath.IsAbs(key) {
		key = filepath.Clean(key)
	}

	if key == "ln" && val == "target" {
		im.useLinkTarget = true
	}

	im.icons[key] = val
}

func (im iconMap) get(f *file) string {
	if val, ok := im.icons[f.path]; ok {
		return val
	}

	if f.IsDir() {
		if val, ok := im.icons[f.Name()+"/"]; ok {
			return val
		}
	}

	var key string

	switch {
	case f.linkState == working && !im.useLinkTarget:
		key = "ln"
	case f.linkState == broken:
		key = "or"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0002 != 0:
		key = "tw"
	case f.IsDir() && f.Mode()&0002 != 0:
		key = "ow"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0:
		key = "st"
	case f.IsDir():
		key = "di"
	case f.Mode()&os.ModeNamedPipe != 0:
		key = "pi"
	case f.Mode()&os.ModeSocket != 0:
		key = "so"
	case f.Mode()&os.ModeCharDevice != 0:
		key = "cd"
	case f.Mode()&os.ModeDevice != 0:
		key = "bd"
	case f.Mode()&os.ModeSetuid != 0:
		key = "su"
	case f.Mode()&os.ModeSetgid != 0:
		key = "sg"
	case f.Mode()&0111 != 0:
		key = "ex"
	}

	if val, ok := im.icons[key]; ok {
		return val
	}

	if val, ok := im.icons[f.Name()+"*"]; ok {
		return val
	}

	if val, ok := im.icons["*"+f.Name()]; ok {
		return val
	}

	if val, ok := im.icons[filepath.Base(f.Name())+".*"]; ok {
		return val
	}

	if val, ok := im.icons["*"+strings.ToLower(f.ext)]; ok {
		return val
	}

	if val, ok := im.icons["fi"]; ok {
		return val
	}

	return " "
}
