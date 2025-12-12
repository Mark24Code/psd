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
	case "linear_light", "lLit":
		return blendLinearLight
	case "color", "colr":
		return blendColor
	case "vivid_light", "vLit":
		return blendVividLight
	case "pin_light", "pLit":
		return blendPinLight
	case "hard_mix", "hMix":
		return blendHardMix
	case "hue", "hue ":
		return blendHue
	case "saturation", "sat ":
		return blendSaturation
	case "luminosity", "lum ":
		return blendLuminosity
	case "subtract", "fsub":
		return blendSubtract
	case "divide", "fdiv":
		return blendDivide
	case "dissolve", "diss":
		return blendDissolve
	case "darker_color", "dkCl":
		return blendDarkerColor
	case "lighter_color", "lgCl":
		return blendLighterColor
	case "passthru", "pass":
		return blendNormal // passthru is same as normal
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

// blendLinearLight performs linear light blend mode
// EXACTLY matches Ruby psd.rb compose.rb:197-213
func blendLinearLight(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8, sa8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8), uint8(sa>>8)
	dr8, dg8, db8, da8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8), uint8(da>>8)

	// Ruby: return apply_opacity(fg, opacity) if fully_transparent?(bg)
	if da8 == 0 {
		// apply_opacity formula: (color & 0xffffff00) | ((color & 0x000000ff) * opacity / 255)
		outA := uint8((uint32(sa8) * uint32(opacity)) / 255)
		return color.RGBA{sr8, sg8, sb8, outA}
	}

	// Ruby: return bg if fully_transparent?(fg)
	if sa8 == 0 {
		return color.RGBA{dr8, dg8, db8, da8}
	}

	// Ruby: calculate_alphas
	// src_alpha = a(fg) * opacity >> 8
	srcAlpha := (uint32(sa8) * uint32(opacity)) >> 8

	// dst_alpha = a(bg)
	dstAlpha := uint32(da8)

	// mix_alpha = (src_alpha << 8) / (src_alpha + ((256 - src_alpha) * dst_alpha >> 8))
	mixAlpha := (srcAlpha << 8) / (srcAlpha + ((256-srcAlpha)*dstAlpha)>>8)

	// dst_alpha = dst_alpha + ((256 - dst_alpha) * src_alpha >> 8)
	outAlpha := dstAlpha + ((256-dstAlpha)*srcAlpha)>>8

	// Ruby: calculate_foreground for linear_light
	// if b < 255 then [f * f / (255 - b), 255].min else 255
	blendR := linearLightChannel(sr8, dr8)
	blendG := linearLightChannel(sg8, dg8)
	blendB := linearLightChannel(sb8, db8)

	// Ruby: blend_channel(bg, fg, mix_alpha)
	// Formula: ((bg << 8) + (fg - bg) * alpha) >> 8
	newR := ((uint32(dr8) << 8) + (uint32(blendR)-uint32(dr8))*mixAlpha) >> 8
	newG := ((uint32(dg8) << 8) + (uint32(blendG)-uint32(dg8))*mixAlpha) >> 8
	newB := ((uint32(db8) << 8) + (uint32(blendB)-uint32(db8))*mixAlpha) >> 8

	return color.RGBA{
		uint8(newR),
		uint8(newG),
		uint8(newB),
		uint8(outAlpha),
	}
}

func linearLightChannel(f, b uint8) uint8 {
	// Ruby: if b < 255 then [f * f / (255 - b), 255].min else 255
	if b < 255 {
		// f * f / (255 - b)
		result := uint32(f) * uint32(f) / uint32(255-b)
		if result > 255 {
			return 255
		}
		return uint8(result)
	}
	return 255
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

// rgbToHSL converts RGB to HSL color space
// H: 0-360, S: 0-1, L: 0-1
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	// Normalize to 0-1
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	// Lightness
	l = (max + min) / 2.0

	if delta == 0 {
		// Achromatic (gray)
		return 0, 0, l
	}

	// Saturation
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2.0 - max - min)
	}

	// Hue
	switch max {
	case rf:
		h = ((gf - bf) / delta)
		if gf < bf {
			h += 6.0
		}
	case gf:
		h = ((bf - rf) / delta) + 2.0
	case bf:
		h = ((rf - gf) / delta) + 4.0
	}
	h *= 60.0

	return h, s, l
}

// hslToRGB converts HSL to RGB color space
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		// Achromatic
		val := uint8(l * 255)
		return val, val, val
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}
	p := 2.0*l - q

	// Helper function for RGB channels
	hueToRGB := func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		if t < 1.0/6.0 {
			return p + (q-p)*6.0*t
		}
		if t < 0.5 {
			return q
		}
		if t < 2.0/3.0 {
			return p + (q-p)*(2.0/3.0-t)*6.0
		}
		return p
	}

	h /= 360.0
	r = uint8(hueToRGB(p, q, h+1.0/3.0) * 255)
	g = uint8(hueToRGB(p, q, h) * 255)
	b = uint8(hueToRGB(p, q, h-1.0/3.0) * 255)

	return r, g, b
}

// blendColor performs color blend mode (HSL-based)
// Takes hue and saturation from source, luminosity from destination
func blendColor(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Apply layer opacity
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{uint8(dr >> 8), uint8(dg >> 8), uint8(db >> 8), uint8(da >> 8)}
	}

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// If destination is fully transparent, just return source
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Convert to HSL
	srcH, srcS, _ := rgbToHSL(sr8, sg8, sb8)
	_, _, dstL := rgbToHSL(dr8, dg8, db8)

	// Combine: source hue/saturation + destination luminosity
	blendR, blendG, blendB := hslToRGB(srcH, srcS, dstL)

	// Alpha composite the result
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	dr8 = uint8(dr >> 8)
	dg8 = uint8(dg >> 8)
	db8 = uint8(db >> 8)

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendVividLight performs vivid light blend mode
// Ruby formula: if (f < 255) then [(b * b / (255 - f) + f * f / (255 - b)) >> 1, 255].min else b
func blendVividLight(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Vivid Light blend using Ruby's formula
	vividLightChannel := func(s, d uint8) uint8 {
		// Handle edge cases to prevent division by zero
		if s == 255 && d == 255 {
			return 255
		}
		if s == 255 {
			return d
		}
		if d == 255 {
			return 255
		}

		// (b * b / (255 - f) + f * f / (255 - b)) >> 1
		term1 := uint32(d) * uint32(d) / uint32(255-s)
		term2 := uint32(s) * uint32(s) / uint32(255-d)
		result := (term1 + term2) >> 1
		if result > 255 {
			return 255
		}
		return uint8(result)
	}

	blendR := vividLightChannel(sr8, dr8)
	blendG := vividLightChannel(sg8, dg8)
	blendB := vividLightChannel(sb8, db8)

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendPinLight performs pin light blend mode
// Ruby formula: if (f >= 128) then [b, (f - 128) * 2].max else [b, f * 2].min
func blendPinLight(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Pin Light blend using Ruby's formula
	pinLightChannel := func(s, d uint8) uint8 {
		if s >= 128 {
			// max(b, (f - 128) * 2)
			temp := (int(s) - 128) * 2
			if int(d) > temp {
				return d
			}
			return uint8(temp)
		}
		// min(b, f * 2)
		temp := int(s) * 2
		if int(d) < temp {
			return d
		}
		if temp > 255 {
			return 255
		}
		return uint8(temp)
	}

	blendR := pinLightChannel(sr8, dr8)
	blendG := pinLightChannel(sg8, dg8)
	blendB := pinLightChannel(sb8, db8)

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendHardMix performs hard mix blend mode
// Ruby formula: (b + f <= 255) ? 0 : 255
func blendHardMix(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Hard Mix blend using Ruby's formula
	hardMixChannel := func(s, d uint8) uint8 {
		if int(d)+int(s) <= 255 {
			return 0
		}
		return 255
	}

	blendR := hardMixChannel(sr8, dr8)
	blendG := hardMixChannel(sg8, dg8)
	blendB := hardMixChannel(sb8, db8)

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendHue performs hue blend mode (HSL-based)
// Takes hue from source, saturation and luminosity from destination
func blendHue(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Apply layer opacity
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{uint8(dr >> 8), uint8(dg >> 8), uint8(db >> 8), uint8(da >> 8)}
	}

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// If destination is fully transparent, just return source
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Convert to HSL
	srcH, _, _ := rgbToHSL(sr8, sg8, sb8)
	_, dstS, dstL := rgbToHSL(dr8, dg8, db8)

	// Combine: source hue + destination saturation/luminosity
	blendR, blendG, blendB := hslToRGB(srcH, dstS, dstL)

	// Alpha composite the result
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	dr8 = uint8(dr >> 8)
	dg8 = uint8(dg >> 8)
	db8 = uint8(db >> 8)

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendSaturation performs saturation blend mode (HSL-based)
// Takes saturation from source, hue and luminosity from destination
func blendSaturation(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Apply layer opacity
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{uint8(dr >> 8), uint8(dg >> 8), uint8(db >> 8), uint8(da >> 8)}
	}

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// If destination is fully transparent, just return source
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Convert to HSL
	_, srcS, _ := rgbToHSL(sr8, sg8, sb8)
	dstH, _, dstL := rgbToHSL(dr8, dg8, db8)

	// Combine: source saturation + destination hue/luminosity
	blendR, blendG, blendB := hslToRGB(dstH, srcS, dstL)

	// Alpha composite the result
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	dr8 = uint8(dr >> 8)
	dg8 = uint8(dg >> 8)
	db8 = uint8(db >> 8)

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendLuminosity performs luminosity blend mode (HSL-based)
// Takes luminosity from source, hue and saturation from destination
func blendLuminosity(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Apply layer opacity
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{uint8(dr >> 8), uint8(dg >> 8), uint8(db >> 8), uint8(da >> 8)}
	}

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// If destination is fully transparent, just return source
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Convert to HSL
	_, _, srcL := rgbToHSL(sr8, sg8, sb8)
	dstH, dstS, _ := rgbToHSL(dr8, dg8, db8)

	// Combine: source luminosity + destination hue/saturation
	blendR, blendG, blendB := hslToRGB(dstH, dstS, srcL)

	// Alpha composite the result
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	dr8 = uint8(dr >> 8)
	dg8 = uint8(dg >> 8)
	db8 = uint8(db >> 8)

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendSubtract performs subtract blend mode
// Formula: max(dst - src, 0)
func blendSubtract(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Subtract blend: max(dst - src, 0)
	subtractChannel := func(s, d uint8) uint8 {
		if int(d) < int(s) {
			return 0
		}
		return d - s
	}

	blendR := subtractChannel(sr8, dr8)
	blendG := subtractChannel(sg8, dg8)
	blendB := subtractChannel(sb8, db8)

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendDivide performs divide blend mode
// Formula: min(dst / src * 255, 255), special handling for src == 0
func blendDivide(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Divide blend: min((dst / src) * 255, 255)
	divideChannel := func(s, d uint8) uint8 {
		if s == 0 {
			return 255
		}
		result := (uint32(d) * 255) / uint32(s)
		if result > 255 {
			return 255
		}
		return uint8(result)
	}

	blendR := divideChannel(sr8, dr8)
	blendG := divideChannel(sg8, dg8)
	blendB := divideChannel(sb8, db8)

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendDissolve performs dissolve blend mode
// Note: Dissolve requires random dithering, which Ruby also doesn't fully implement
// For now, fall back to normal blend mode
func blendDissolve(src, dst color.Color, opacity uint8) color.RGBA {
	// Dissolve is a special mode that requires random dithering per pixel
	// Ruby's compose.rb also doesn't implement this, falling back to normal
	return blendNormal(src, dst, opacity)
}

// blendDarkerColor performs darker color blend mode
// Compares total luminosity and picks the darker color
func blendDarkerColor(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Calculate luminosity (approximation: 0.299*R + 0.587*G + 0.114*B)
	srcLum := uint32(sr8)*299 + uint32(sg8)*587 + uint32(sb8)*114
	dstLum := uint32(dr8)*299 + uint32(dg8)*587 + uint32(db8)*114

	var blendR, blendG, blendB uint8
	if srcLum < dstLum {
		blendR, blendG, blendB = sr8, sg8, sb8
	} else {
		blendR, blendG, blendB = dr8, dg8, db8
	}

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}

// blendLighterColor performs lighter color blend mode
// Compares total luminosity and picks the lighter color
func blendLighterColor(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Convert to 8-bit
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)
	dr8, dg8, db8 := uint8(dr>>8), uint8(dg>>8), uint8(db>>8)

	// Apply layer opacity to source alpha
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{dr8, dg8, db8, uint8(da >> 8)}
	}
	if da == 0 {
		return color.RGBA{sr8, sg8, sb8, uint8(alpha)}
	}

	// Calculate luminosity (approximation: 0.299*R + 0.587*G + 0.114*B)
	srcLum := uint32(sr8)*299 + uint32(sg8)*587 + uint32(sb8)*114
	dstLum := uint32(dr8)*299 + uint32(dg8)*587 + uint32(db8)*114

	var blendR, blendG, blendB uint8
	if srcLum > dstLum {
		blendR, blendG, blendB = sr8, sg8, sb8
	} else {
		blendR, blendG, blendB = dr8, dg8, db8
	}

	// Alpha compositing
	outAlpha := alpha + (da*(255-alpha))/255
	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	outRed := (uint32(blendR)*alpha + uint32(dr8)*da*(255-alpha)/255) / outAlpha
	outGreen := (uint32(blendG)*alpha + uint32(dg8)*da*(255-alpha)/255) / outAlpha
	outBlue := (uint32(blendB)*alpha + uint32(db8)*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
}
