package yam

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"unsafe"
	"yam/yrender"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth      = 1280
	winHeight     = 720
	minWarpFactor = 0.1
	numStars      = 300
	centerX       = winWidth / 2
	centerY       = winHeight / 2
	stride        = 4
)

type coords struct {
	x, y float64
}

type star struct {
	pos, vel   coords
	brightness byte
}

type stars struct {
	stars []star
}

func randFloat64(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c byte, pixels []byte) {
	index := (y*winWidth + x) * stride

	if index < len(pixels)-stride && index >= 0 {
		pixels[index] = c
		pixels[index+1] = c
		pixels[index+2] = c
	}
}

func newStar() star {
	angle := randFloat64(float64(-3.14), float64(3.14))
	speed := 255 * math.Pow(randFloat64(float64(0.3), float64(1.0)), 2)
	dx := math.Cos(angle)
	dy := math.Sin(angle)

	d := rand.Intn(100) + 25 //+ traillength
	// pos = centerx + dx * d, centery + dy * d
	pos := coords{
		x: centerX + dx*float64(d),
		y: centerY + dy*float64(d),
	}
	// vel = speed * dx, speed * dy
	vel := coords{
		x: speed * dx,
		y: speed * dy,
	}

	s := star{
		pos:        pos,
		vel:        vel,
		brightness: 0,
	}
	return s
}

func (s *stars) update(dt float64) {
	for i := 0; i < len(s.stars); i++ {
		newPosX := (s.stars[i].pos.x + (s.stars[i].vel.x * minWarpFactor)) //* dt
		newPosY := s.stars[i].pos.y + (s.stars[i].vel.y * minWarpFactor)   //*dt
		if newPosX > winWidth || newPosY > winHeight || newPosX < 0 || newPosY < 0 {
			s.stars[i] = newStar()
		} else {
			s.stars[i].pos.x = newPosX
			s.stars[i].pos.y = newPosY

			// # Grow brighter
			// s.brightness = min(s.brightness + warp_factor * 200 * dt, s.speed)
			if s.stars[i].brightness < 255 {
				s.stars[i].brightness += 40
			}
		}
	}

}

func (s *stars) draw(pixels []byte) {
	for i := 0; i < len(s.stars); i++ {
		if int(s.stars[i].pos.x) >= 0 {
			setPixel(int(s.stars[i].pos.x), int(s.stars[i].pos.y), s.stars[i].brightness, pixels)
		}
	}
}

func starTest() {
	fmt.Println("welcome to yam!")
	w, err := yrender.CreateSDLRenderer("test", winWidth, winHeight)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	tex, err := w.Renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		panic(err)
	}
	defer tex.Destroy()
	var elapsedTime float64
	pixels := make([]byte, winWidth*winHeight*4)
	starField := make([]star, numStars)
	all := &stars{}

	for i := 0; i < len(starField); i++ {
		all.stars = append(all.stars, newStar())
	}
	for i := 0; i < 2000; i++ {
		all.update(0)
		clear(pixels)
	}

	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		all.update(elapsedTime)
		all.draw(pixels)

		tex.Update(nil, unsafe.Pointer(&pixels[0]), winWidth*4)
		w.Renderer.Copy(tex, nil, nil)
		w.Renderer.Present()
		clear(pixels)
		elapsedTime = float64(time.Since(frameStart).Seconds() * 1000)
		if elapsedTime < 7 {
			sdl.Delay(7 - uint32(elapsedTime))
			elapsedTime = float64(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
