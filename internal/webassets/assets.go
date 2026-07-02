package webassets

import "embed"

//go:embed static static/* static/css/* static/js/* static/img/* static/img/cards/* static/img/bosses/*
var Assets embed.FS
