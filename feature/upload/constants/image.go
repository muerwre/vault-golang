package constants

const (
	FileMimeJpeg  string = "image/jpeg"
	FileMimePng   string = "image/png"
	FileMimeGif   string = "image/gif"
	FileMimeMpeg3 string = "audio/mpeg3"
	FileMimeMpeg  string = "audio/mpeg"
	FileMimeMp3   string = "audio/mp3"
)

const (
	ImagePreset1600      string = "1600"
	ImagePreset600       string = "600"
	ImagePreset300       string = "300"
	ImagePresetAvatar    string = "avatar"
	ImagePresetCover     string = "cover"
	ImagePresetSmallHero string = "small_hero"
)

const (
	FileTypeImage string = "image"
	FileTypeAudio string = "audio"
)

var FileTypeToMime = map[string][]string{
	FileTypeImage: {FileMimeJpeg, FileMimePng, FileMimeGif},
	FileTypeAudio: {FileMimeMp3, FileMimeMpeg, FileMimeMpeg3},
}
