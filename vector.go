package main

import (
	"fmt"
	"math"
)

type Vector struct {
	X, Y float64
}

func NewVector(x, y float64) Vector {
	return Vector{x, y}
}

func (v Vector) Angle() float64 {
	return math.Atan2(v.Y, v.X)
}

// Implement helper functions for Vector
func (v Vector) Add(u Vector) Vector {
	return Vector{v.X + u.X, v.Y + u.Y}
}

func (v Vector) Sub(u Vector) Vector {
	return Vector{v.X - u.X, v.Y - u.Y}
}

func (v Vector) Scaled(s float64) Vector {
	return Vector{v.X * s, v.Y * s}
}

func (v Vector) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector) Unit() Vector {
	return v.Scaled(1 / v.Len())
}

func (v Vector) String() string {
	return fmt.Sprintf("(%f, %f)", v.X, v.Y)
}
