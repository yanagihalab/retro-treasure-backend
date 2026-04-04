package webassets

import "embed"

//go:embed static static/* static/css/* static/js/*
var Assets embed.FS
