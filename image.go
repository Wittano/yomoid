package main

import (
	"cmp"
	"context"
	"errors"
	"image"
	"io"
	"math"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var ErrMainColorNotFound = errors.New("poll: failed detect main color")

var client = http.Client{
	Timeout: time.Second * 5,
}

var onlyPngUrlRegex = regexp.MustCompile("\\.[a-z]{3,4}\\?")

func downloadImage(url string) (io.ReadCloser, error) {
	url = onlyPngUrlRegex.ReplaceAllString(url, ".png?")

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func imageMainColor(ctx context.Context, reader io.Reader) (uint32, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return 0, err
	}

	var (
		m      sync.Mutex
		wg     sync.WaitGroup
		colors = make(map[uint32]int)
		radius = img.Bounds().Dx() / 2
	)
	for x := 0; x < img.Bounds().Dx(); x++ {
		wg.Add(1)

		go func(dx int) {
			defer wg.Done()

			for y := 0; y < img.Bounds().Dy(); y++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if circleFormula(x, y, radius) > 0 {
					continue
				}

				r, g, b, _ := img.At(x, y).RGBA()
				rgb := ((r >> 8) << 16) | ((g >> 8) << 8) | (b >> 8)
				if rgb == 0 {
					continue
				}

				m.Lock()
				if val, ok := colors[rgb]; ok {
					colors[rgb] = val + 1
				} else {
					colors[rgb] = 1
				}
				m.Unlock()
			}
		}(x)
	}

	wg.Wait()

	var (
		color         uint32
		maxColorCount float64 = 0
	)
	for c, count := range colors {
		select {
		case <-ctx.Done():
			return 0, context.Canceled
		default:
		}

		newMax := math.Max(maxColorCount, float64(count))
		if newMax != maxColorCount {
			maxColorCount = newMax
			color = c
		}
	}

	if color == 0 {
		return 0, ErrMainColorNotFound
	}

	return color, nil
}

func circleFormula(x, y, r int) int {
	circle := (x-r)*(x-r) + (y-r)*(y-r)

	return cmp.Compare(circle, r*r)
}
