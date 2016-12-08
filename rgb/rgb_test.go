package rgb_test

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/jpeg"
	"io/ioutil"
	"testing"

	"github.com/runningwild/argus/rgb"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEncodeAndDecode(t *testing.T) {
	Convey("can draw onto rgb.Image", t, func() {
		data, err := ioutil.ReadFile("testdata/buttons.jpg")
		So(err, ShouldBeNil)
		src, _, err := image.Decode(bytes.NewBuffer(data))
		So(err, ShouldBeNil)
		dst := rgb.Make(src.Bounds())
		So(dst.Bounds().Dx(), ShouldEqual, src.Bounds().Dx())
		So(dst.Bounds().Dy(), ShouldEqual, src.Bounds().Dy())
		draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Over)

		// Check part of the image, we don't need to go crazy
		for y := 0; y < 100 && y < src.Bounds().Dy(); y++ {
			for x := 0; x < 100 && x < src.Bounds().Dx(); x++ {
				sr, sg, sb, _ := src.At(x, y).RGBA()
				dr, dg, db, _ := dst.At(x, y).RGBA()
				So(dr>>8, ShouldEqual, sr>>8)
				So(dg>>8, ShouldEqual, sg>>8)
				So(db>>8, ShouldEqual, sb>>8)
			}
		}
	})
}
