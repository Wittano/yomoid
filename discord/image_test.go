package discord

import (
	"context"
	"github.com/wittano/yomoid/logger"
	"os"
	"testing"
	"time"
)

func TestDownloadImageAndSearchMainColor(t *testing.T) {
	img, err := downloadImage("https://cdn.discordapp.com/avatars/299551004628615178/338422b3211015675aa87d130bba6aec.png?size=128")
	if err != nil {
		t.Fatal(err)
	}
	defer logger.LogCloser(img)

	got, err := imageMainColor(context.Background(), img)
	if err != nil {
		t.Fatal(err)
	}

	const exp = 0xffe8be
	if got != exp {
		t.Fatalf("invalid main image color. got #%x; want #%x", got, exp)
	}
}

func TestImageMainColor(t *testing.T) {
	data := map[string]uint32{
		"big-img.png":    0x3498db,
		"small.png":      0x68aa53,
		"large-icon.jpg": 0x00ab93,
	}

	for filename, exp := range data {
		t.Run(filename, func(t *testing.T) {
			f, err := os.Open("testdata/" + filename)
			if err != nil {
				t.Fatal(err)
			}
			defer logger.LogCloser(f)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			got, err := imageMainColor(ctx, f)
			if err != nil {
				t.Fatal(err)
			}

			if got != exp {
				t.Fatalf("invalid image main color result. Expected: #%x, got: #%x", exp, got)
			}
		})
	}
}
