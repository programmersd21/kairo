package render

// Clamp01 constrains a value to the [0, 1] range.
// Use this to guarantee easing output never overshoots, which
// prevents index-out-of-bounds and visual flicker in text animations.
func Clamp01(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t
}

// Linear is a simple linear easing (identity function).
func Linear(t float64) float64 {
	return Clamp01(t)
}

// EaseOutQuad starts fast and decelerates. Monotonic, no overshoot.
func EaseOutQuad(t float64) float64 {
	t = Clamp01(t)
	return t * (2 - t)
}

// EaseOutCubic is the primary animation easing for production UI.
// Smooth deceleration with a natural, polished feel.
// Strictly monotonic — never overshoots 1.0.
func EaseOutCubic(t float64) float64 {
	t = Clamp01(t)
	t -= 1
	return t*t*t + 1
}

// EaseInOutCubic accelerates then decelerates.
// Smooth S-curve suitable for longer-duration transitions.
func EaseInOutCubic(t float64) float64 {
	t = Clamp01(t)
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - (-2*t+2)*(-2*t+2)*(-2*t+2)/2
}
