package cs2log

import "fmt"

// Position and velocity helpers use a prefix plus X/Y/Z field names. Keeping
// the naming rule here lets definitions compose larger field groups without
// repeating the parser logic in each event builder.
func positionFromMatches(matches Matches, prefix string) (Position, error) {
	x, err := matches.Int(prefix + "X")
	if err != nil {
		return Position{}, fmt.Errorf("%sX: %w", prefix, err)
	}
	y, err := matches.Int(prefix + "Y")
	if err != nil {
		return Position{}, fmt.Errorf("%sY: %w", prefix, err)
	}
	z, err := matches.Int(prefix + "Z")
	if err != nil {
		return Position{}, fmt.Errorf("%sZ: %w", prefix, err)
	}
	return Position{X: x, Y: y, Z: z}, nil
}

func positionFloatFromMatches(matches Matches, prefix string) (PositionFloat, error) {
	x, err := matches.Float64(prefix + "X")
	if err != nil {
		return PositionFloat{}, fmt.Errorf("%sX: %w", prefix, err)
	}
	y, err := matches.Float64(prefix + "Y")
	if err != nil {
		return PositionFloat{}, fmt.Errorf("%sY: %w", prefix, err)
	}
	z, err := matches.Float64(prefix + "Z")
	if err != nil {
		return PositionFloat{}, fmt.Errorf("%sZ: %w", prefix, err)
	}
	return PositionFloat{X: x, Y: y, Z: z}, nil
}

func velocityFromMatches(matches Matches, prefix string) (Velocity, error) {
	x, err := matches.Float64(prefix + "X")
	if err != nil {
		return Velocity{}, fmt.Errorf("%sX: %w", prefix, err)
	}
	y, err := matches.Float64(prefix + "Y")
	if err != nil {
		return Velocity{}, fmt.Errorf("%sY: %w", prefix, err)
	}
	z, err := matches.Float64(prefix + "Z")
	if err != nil {
		return Velocity{}, fmt.Errorf("%sZ: %w", prefix, err)
	}
	return Velocity{X: x, Y: y, Z: z}, nil
}

func positionFields(prefix string) []field {
	return []field{
		{Name: prefix + "X", Type: "int"},
		{Name: prefix + "Y", Type: "int"},
		{Name: prefix + "Z", Type: "int"},
	}
}

func positionFloatFields(prefix string) []field {
	return []field{
		{Name: prefix + "X", Type: "float"},
		{Name: prefix + "Y", Type: "float"},
		{Name: prefix + "Z", Type: "float"},
	}
}
