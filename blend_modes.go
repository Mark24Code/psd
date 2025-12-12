package psd

import (
	"image/color"
	"math"
)

// BlendFunc is a function that blends two colors
type BlendFunc func(src, dst color.Color, opacity uint8) color.RGBA

// GetBlendFunc returns the appropriate blend function for a blend mode
func GetBlendFunc(blendMode string) BlendFunc {
	switch blendMode {
	case "normal", "norm":
		return blendNormal
	case "multiply", "mul ":
		return blendMultiply
	case "screen", "scrn":
		return blendScreen
	case "overlay", "over":
		return blendOverlay
	case "darken", "dark":
		return blendDarken
	case "lighten", "lite":
		return blendLighten
	case "color_dodge", "div ":
		return blendColorDodge
	case "color_burn", "idiv":
		return blendColorBurn
	case "hard_light", "hLit":
		return blendHardLight
	case "soft_light", "sLit":
		return blendSoftLight
	case "difference", "diff":
		return blendDifference
	case "exclusion", "smud":
		return blendExclusion
	case "linear_dodge", "lddg":
		return blendLinearDodge
	case "linear_burn", "lbrn":
		return blendLinearBurn
	default:
		// Default to normal for unknown modes
		return blendNormal
	}
}

// blendNormal performs normal blend mode (source over)
func blendNormal(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Apply layer opacity
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{uint8(dr >> 8), uint8(dg >> 8), uint8(db >> 8), uint8(da >> 8)}
	}

	if alpha == 255 && da == 0 {
		return color.RGBA{uint8(sr >> 8), uint8(sg >> 8), uint8(sb >> 8), uint8(alpha)}
	}

	// Alpha compositing: C = (Cs * As + Cd * Ad * (1 - As)) / Ao
	// where Ao = As + Ad * (1 - As)
	outAlpha := alpha + (da*(255-alpha))/255

	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	// Convert from 16-bit to 8-bit color space
	sr8, sg8, sb8 := sr>>8, sg>>8, sb>>8
	dr8, dg8, db8 := dr>>8, dg>>8, db>>8

	// Blend colors
	outRed := (sr8*alpha + dr8*da*(255-alpha)/255) / outAlpha
	outGreen := (sg8*alpha + dg8*da*(255-alpha)/255) / outAlpha
	outBlue := (sb8*alpha + db8*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendMultiply performs multiply blend mode
func blendMultiply(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Multiply blend: C = Cs * Cd
	blendR := sr * dr
	blendG := sg * dg
	blendB := sb * db

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendScreen performs screen blend mode
func blendScreen(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Screen blend: C = 1 - (1 - Cs) * (1 - Cd)
	blendR := 1.0 - (1.0-sr)*(1.0-dr)
	blendG := 1.0 - (1.0-sg)*(1.0-dg)
	blendB := 1.0 - (1.0-sb)*(1.0-db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendOverlay performs overlay blend mode
func blendOverlay(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Overlay: C = (Cd < 0.5) ? (2 * Cs * Cd) : (1 - 2 * (1 - Cs) * (1 - Cd))
	blendR := overlayChannel(sr, dr)
	blendG := overlayChannel(sg, dg)
	blendB := overlayChannel(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

func overlayChannel(s, d float64) float64 {
	if d < 0.5 {
		return 2.0 * s * d
	}
	return 1.0 - 2.0*(1.0-s)*(1.0-d)
}

// blendDarken performs darken blend mode
func blendDarken(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Darken: C = min(Cs, Cd)
	blendR := math.Min(sr, dr)
	blendG := math.Min(sg, dg)
	blendB := math.Min(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendLighten performs lighten blend mode
func blendLighten(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Lighten: C = max(Cs, Cd)
	blendR := math.Max(sr, dr)
	blendG := math.Max(sg, dg)
	blendB := math.Max(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendColorDodge performs color dodge blend mode
func blendColorDodge(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Color Dodge: C = Cd / (1 - Cs)
	blendR := colorDodgeChannel(sr, dr)
	blendG := colorDodgeChannel(sg, dg)
	blendB := colorDodgeChannel(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

func colorDodgeChannel(s, d float64) float64 {
	if s >= 1.0 {
		return 1.0
	}
	result := d / (1.0 - s)
	if result > 1.0 {
		return 1.0
	}
	return result
}

// blendColorBurn performs color burn blend mode
func blendColorBurn(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Color Burn: C = 1 - (1 - Cd) / Cs
	blendR := colorBurnChannel(sr, dr)
	blendG := colorBurnChannel(sg, dg)
	blendB := colorBurnChannel(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

func colorBurnChannel(s, d float64) float64 {
	if s <= 0.0 {
		return 0.0
	}
	result := 1.0 - (1.0-d)/s
	if result < 0.0 {
		return 0.0
	}
	return result
}

// blendHardLight performs hard light blend mode
func blendHardLight(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Hard Light: C = (Cs < 0.5) ? (2 * Cs * Cd) : (1 - 2 * (1 - Cs) * (1 - Cd))
	blendR := hardLightChannel(sr, dr)
	blendG := hardLightChannel(sg, dg)
	blendB := hardLightChannel(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

func hardLightChannel(s, d float64) float64 {
	if s < 0.5 {
		return 2.0 * s * d
	}
	return 1.0 - 2.0*(1.0-s)*(1.0-d)
}

// blendSoftLight performs soft light blend mode
func blendSoftLight(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Soft Light (Pegtop formula): C = (1 - 2 * Cs) * Cd^2 + 2 * Cs * Cd
	blendR := softLightChannel(sr, dr)
	blendG := softLightChannel(sg, dg)
	blendB := softLightChannel(sb, db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

func softLightChannel(s, d float64) float64 {
	return (1.0-2.0*s)*d*d + 2.0*s*d
}

// blendDifference performs difference blend mode
func blendDifference(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Difference: C = |Cs - Cd|
	blendR := math.Abs(sr - dr)
	blendG := math.Abs(sg - dg)
	blendB := math.Abs(sb - db)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendExclusion performs exclusion blend mode
func blendExclusion(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Exclusion: C = Cs + Cd - 2 * Cs * Cd
	blendR := sr + dr - 2.0*sr*dr
	blendG := sg + dg - 2.0*sg*dg
	blendB := sb + db - 2.0*sb*db

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendLinearDodge performs linear dodge blend mode (same as Add)
func blendLinearDodge(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Linear Dodge (Add): C = Cs + Cd
	blendR := math.Min(sr+dr, 1.0)
	blendG := math.Min(sg+dg, 1.0)
	blendB := math.Min(sb+db, 1.0)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// blendLinearBurn performs linear burn blend mode
func blendLinearBurn(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := toFloat(src)
	dr, dg, db, da := toFloat(dst)

	// Linear Burn: C = Cs + Cd - 1
	blendR := math.Max(sr+dr-1.0, 0.0)
	blendG := math.Max(sg+dg-1.0, 0.0)
	blendB := math.Max(sb+db-1.0, 0.0)

	return applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB, opacity)
}

// Helper functions

// toFloat converts color to float64 values [0.0, 1.0]
func toFloat(c color.Color) (r, g, b, a float64) {
	r32, g32, b32, a32 := c.RGBA()
	r = float64(r32) / 65535.0
	g = float64(g32) / 65535.0
	b = float64(b32) / 65535.0
	a = float64(a32) / 65535.0
	return
}

// applyBlend applies the blended colors with opacity and alpha compositing
func applyBlend(sr, sg, sb, sa, dr, dg, db, da, blendR, blendG, blendB float64, opacity uint8) color.RGBA {
	// Apply layer opacity
	alpha := float64(opacity) / 255.0 * sa

	if alpha == 0 {
		return color.RGBA{
			uint8(dr * 255),
			uint8(dg * 255),
			uint8(db * 255),
			uint8(da * 255),
		}
	}

	// Alpha compositing
	outAlpha := alpha + da*(1.0-alpha)

	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	// Composite the blended color
	outRed := (blendR*alpha + dr*da*(1.0-alpha)) / outAlpha
	outGreen := (blendG*alpha + dg*da*(1.0-alpha)) / outAlpha
	outBlue := (blendB*alpha + db*da*(1.0-alpha)) / outAlpha

	return color.RGBA{
		uint8(clamp(outRed * 255.0)),
		uint8(clamp(outGreen * 255.0)),
		uint8(clamp(outBlue * 255.0)),
		uint8(clamp(outAlpha * 255.0)),
	}
}

// clamp clamps a value between 0 and 255
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
