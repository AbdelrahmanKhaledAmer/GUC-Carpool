package DirectionsAPI

import (
	"context"
	"log"

	strip "github.com/grokify/html-strip-tags-go"
	"googlemaps.github.io/maps"
)

//GetRoute : return a string with the instructions to follow to reach the destination
func GetRoute(from string, to string) (string, error) {

	c, err := maps.NewClient(maps.WithAPIKey("AIzaSyCk9fnMYUnWm33Ce4JxCcITloHnj3WncLU"))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
		return "", err
	}
	rm := &maps.DirectionsRequest{
		Origin:      from,
		Destination: to,
	}
	route, _, err := c.Directions(context.Background(), rm)
	if err != nil {
		return "", err
	}

	routeInstructions := ""
	for i1 := 0; i1 < len(route); i1++ {
		for i2 := 0; i2 < len(route[i1].Legs); i2++ {
			for i3 := 0; i3 < len(route[i1].Legs[i2].Steps); i3++ {
				routeInstructions += strip.StripTags(route[i1].Legs[i2].Steps[i3].HTMLInstructions) + "\n"
			}
		}
	}
	return routeInstructions, nil
}
