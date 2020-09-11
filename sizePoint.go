package main

import (
	"image"
)

type sizePoint struct {
	X, Y int
}

func CreatePoint(p image.Point) sizePoint {
	return sizePoint{p.X, p.Y}
}

func (o sizePoint) GetMin() int {
	if o.X < o.Y {
		return o.X
	}
	return o.Y
}

func (o sizePoint) GetMax() int {
	if o.X > o.Y {
		return o.X
	}
	return o.Y
}

func (o sizePoint) Equal() bool {
	return o.X == o.Y
}
