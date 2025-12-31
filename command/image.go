package command

import (
	"net/http"
	"slices"

	"github.com/disintegration/imaging"
	"github.com/makinori/mikogo/irc"
	"github.com/makinori/mikogo/ircimage"
)

func handleFunImage(msg *irc.Message, args []string) {
	if len(args) < 2 {
		msg.Client.Send(msg.Where, "usage: [-nodither] <image url>")
		return
	}

	noDither := slices.Contains(args, "-nodither")

	res, err := http.Get(args[len(args)-1])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get image: "+err.Error())
		return
	}
	defer res.Body.Close()

	image, err := imaging.Decode(res.Body)
	if err != nil {
		msg.Client.Send(msg.Where, "failed to decode image: "+err.Error())
		return
	}

	img := imaging.Resize(image, 0, 20, imaging.Lanczos)
	img = imaging.AdjustContrast(img, 10)

	var encodedImg ircimage.HalfBlockImage
	if noDither {
		encodedImg, err = ircimage.ConvertImageWithColorCodesNodither(img)
	} else {
		encodedImg, err = ircimage.ConvertImageWithColorCodesDither(img, 32, 0.8)
	}
	if err != nil {
		msg.Client.Send(msg.Where, "failed to convert image: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, encodedImg.IRC())
}

var CommandFunImage = Command{
	Name:        "image",
	Category:    "fun",
	Description: "display image using formatting",
	Handle:      handleFunImage,
}
