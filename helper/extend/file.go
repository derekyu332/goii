package extend

import "path"

func IsFileImage(filename string) bool {
	ext := path.Ext(filename)

	switch ext {
	case ".png", ".gif", ".jpeg", ".jpg":
		return true
	default:
		return false
	}
}

func IsFileAudio(filename string) bool {
	ext := path.Ext(filename)

	switch ext {
	case ".amr", ".mp3", ".wav", ".wma", "m4a":
		return true
	default:
		return false
	}
}
