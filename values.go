package cs2log

// Entity is a non-player entity token from a CS2 log line.
type Entity struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// Position is an integer world coordinate from a CS2 log line.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// PositionFloat is a floating-point world coordinate from a CS2 log line.
type PositionFloat struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Velocity is a floating-point projectile velocity from a CS2 log line.
type Velocity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Equation describes a money equation in the form A + B = Result.
type Equation struct {
	A      int `json:"a"`
	B      int `json:"b"`
	Result int `json:"result"`
}
