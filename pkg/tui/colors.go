package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ColorPair struct {
	Name  string
	Value lipgloss.Color
}

const (
	White = lipgloss.Color("#fff")
	Black = lipgloss.Color("#000")

	Blue      = lipgloss.Color("#0D6EFD")
	Blue100   = lipgloss.Color("#CFE2FF")
	Blue200   = lipgloss.Color("#9EC5FE")
	Blue300   = lipgloss.Color("#6EA8FE")
	Blue400   = lipgloss.Color("#3D8BFD")
	Blue500   = lipgloss.Color("#0D6EFD")
	Blue600   = lipgloss.Color("#0A58CA")
	Blue700   = lipgloss.Color("#084298")
	Blue800   = lipgloss.Color("#052C65")
	Blue900   = lipgloss.Color("#031633")
	Indigo    = lipgloss.Color("#6610F2")
	Indigo100 = lipgloss.Color("#E0CFFC")
	Indigo200 = lipgloss.Color("#C29FFA")
	Indigo300 = lipgloss.Color("#A370F7")
	Indigo400 = lipgloss.Color("#8540F5")
	Indigo500 = lipgloss.Color("#6610F2")
	Indigo600 = lipgloss.Color("#520DC2")
	Indigo700 = lipgloss.Color("#3D0A91")
	Indigo800 = lipgloss.Color("#290661")
	Indigo900 = lipgloss.Color("#140330")
	Purple    = lipgloss.Color("#6F42C1")
	Purple100 = lipgloss.Color("#E2D9F3")
	Purple200 = lipgloss.Color("#C5B3E6")
	Purple300 = lipgloss.Color("#A98EDA")
	Purple400 = lipgloss.Color("#8C68CD")
	Purple500 = lipgloss.Color("#6F42C1")
	Purple600 = lipgloss.Color("#59359A")
	Purple700 = lipgloss.Color("#432874")
	Purple800 = lipgloss.Color("#2C1A4D")
	Purple900 = lipgloss.Color("#160D27")
	Pink      = lipgloss.Color("#D63384")
	Pink100   = lipgloss.Color("#F7D6E6")
	Pink200   = lipgloss.Color("#EFADCE")
	Pink300   = lipgloss.Color("#E685B5")
	Pink400   = lipgloss.Color("#DE5C9D")
	Pink500   = lipgloss.Color("#D63384")
	Pink600   = lipgloss.Color("#AB296A")
	Pink700   = lipgloss.Color("#801F4F")
	Pink800   = lipgloss.Color("#561435")
	Pink900   = lipgloss.Color("#2B0A1A")
	Red       = lipgloss.Color("#DC3545")
	Red100    = lipgloss.Color("#F8D7DA")
	Red200    = lipgloss.Color("#F1AEB5")
	Red300    = lipgloss.Color("#EA868F")
	Red400    = lipgloss.Color("#E35D6A")
	Red500    = lipgloss.Color("#DC3545")
	Red600    = lipgloss.Color("#B02A37")
	Red700    = lipgloss.Color("#842029")
	Red800    = lipgloss.Color("#58151C")
	Red900    = lipgloss.Color("#2C0B0E")
	Orange    = lipgloss.Color("#FD7E14")
	Orange100 = lipgloss.Color("#FFE5D0")
	Orange200 = lipgloss.Color("#FECBA1")
	Orange300 = lipgloss.Color("#FEB272")
	Orange400 = lipgloss.Color("#FD9843")
	Orange500 = lipgloss.Color("#FD7E14")
	Orange600 = lipgloss.Color("#CA6510")
	Orange700 = lipgloss.Color("#984C0C")
	Orange800 = lipgloss.Color("#653208")
	Orange900 = lipgloss.Color("#331904")
	Yellow    = lipgloss.Color("#FFC107")
	Yellow100 = lipgloss.Color("#FFF3CD")
	Yellow200 = lipgloss.Color("#FFE69C")
	Yellow300 = lipgloss.Color("#FFDA6A")
	Yellow400 = lipgloss.Color("#FFCD39")
	Yellow500 = lipgloss.Color("#FFC107")
	Yellow600 = lipgloss.Color("#CC9A06")
	Yellow700 = lipgloss.Color("#997404")
	Yellow800 = lipgloss.Color("#664D03")
	Yellow900 = lipgloss.Color("#332701")
	Green     = lipgloss.Color("#198754")
	Green100  = lipgloss.Color("#D1E7DD")
	Green200  = lipgloss.Color("#A3CFBB")
	Green300  = lipgloss.Color("#75B798")
	Green400  = lipgloss.Color("#479F76")
	Green500  = lipgloss.Color("#198754")
	Green600  = lipgloss.Color("#146C43")
	Green700  = lipgloss.Color("#0F5132")
	Green800  = lipgloss.Color("#0A3622")
	Green900  = lipgloss.Color("#051B11")
	Teal      = lipgloss.Color("#20C997")
	Teal100   = lipgloss.Color("#D2F4EA")
	Teal200   = lipgloss.Color("#A6E9D5")
	Teal300   = lipgloss.Color("#79DFC1")
	Teal400   = lipgloss.Color("#4DD4AC")
	Teal500   = lipgloss.Color("#20C997")
	Teal600   = lipgloss.Color("#1AA179")
	Teal700   = lipgloss.Color("#13795B")
	Teal800   = lipgloss.Color("#0D503C")
	Teal900   = lipgloss.Color("#06281E")
	Cyan      = lipgloss.Color("#0DCAF0")
	Cyan100   = lipgloss.Color("#CFF4FC")
	Cyan200   = lipgloss.Color("#9EEAF9")
	Cyan300   = lipgloss.Color("#6EDFF6")
	Cyan400   = lipgloss.Color("#3DD5F3")
	Cyan500   = lipgloss.Color("#0DCAF0")
	Cyan600   = lipgloss.Color("#0AA2C0")
	Cyan700   = lipgloss.Color("#087990")
	Cyan800   = lipgloss.Color("#055160")
	Cyan900   = lipgloss.Color("#032830")
	Gray      = lipgloss.Color("#ADB5BD")
	Gray100   = lipgloss.Color("#EFF0F2")
	Gray200   = lipgloss.Color("#DEE1E5")
	Gray300   = lipgloss.Color("#CED3D7")
	Gray400   = lipgloss.Color("#BDC4CA")
	Gray500   = lipgloss.Color("#ADB5BD")
	Gray600   = lipgloss.Color("#8A9197")
	Gray700   = lipgloss.Color("#686D71")
	Gray800   = lipgloss.Color("#45484C")
	Gray900   = lipgloss.Color("#232426")
)

var (
	BlueFamily = []ColorPair{
		{
			Name:  "Blue100",
			Value: Blue100,
		},
		{
			Name:  "Blue200",
			Value: Blue200,
		},
		{
			Name:  "Blue300",
			Value: Blue300,
		},
		{
			Name:  "Blue400",
			Value: Blue400,
		},
		{
			Name:  "Blue500",
			Value: Blue500,
		},
		{
			Name:  "Blue600",
			Value: Blue600,
		},
		{
			Name:  "Blue700",
			Value: Blue700,
		},
		{
			Name:  "Blue800",
			Value: Blue800,
		},
		{
			Name:  "Blue900",
			Value: Blue900,
		},
	}
	IndigoFamily = []ColorPair{
		{
			Name:  "Indigo100",
			Value: Indigo100,
		},
		{
			Name:  "Indigo200",
			Value: Indigo200,
		},
		{
			Name:  "Indigo300",
			Value: Indigo300,
		},
		{
			Name:  "Indigo400",
			Value: Indigo400,
		},
		{
			Name:  "Indigo500",
			Value: Indigo500,
		},
		{
			Name:  "Indigo600",
			Value: Indigo600,
		},
		{
			Name:  "Indigo700",
			Value: Indigo700,
		},
		{
			Name:  "Indigo800",
			Value: Indigo800,
		},
		{
			Name:  "Indigo900",
			Value: Indigo900,
		},
	}
	PurpleFamily = []ColorPair{
		{
			Name:  "Purple100",
			Value: Purple100,
		},
		{
			Name:  "Purple200",
			Value: Purple200,
		},
		{
			Name:  "Purple300",
			Value: Purple300,
		},
		{
			Name:  "Purple400",
			Value: Purple400,
		},
		{
			Name:  "Purple500",
			Value: Purple500,
		},
		{
			Name:  "Purple600",
			Value: Purple600,
		},
		{
			Name:  "Purple700",
			Value: Purple700,
		},
		{
			Name:  "Purple800",
			Value: Purple800,
		},
		{
			Name:  "Purple900",
			Value: Purple900,
		},
	}
	PinkFamily = []ColorPair{
		{
			Name:  "Pink100",
			Value: Pink100,
		},
		{
			Name:  "Pink200",
			Value: Pink200,
		},
		{
			Name:  "Pink300",
			Value: Pink300,
		},
		{
			Name:  "Pink400",
			Value: Pink400,
		},
		{
			Name:  "Pink500",
			Value: Pink500,
		},
		{
			Name:  "Pink600",
			Value: Pink600,
		},
		{
			Name:  "Pink700",
			Value: Pink700,
		},
		{
			Name:  "Pink800",
			Value: Pink800,
		},
		{
			Name:  "Pink900",
			Value: Pink900,
		},
	}
	RedFamily = []ColorPair{
		{
			Name:  "Red100",
			Value: Red100,
		},
		{
			Name:  "Red200",
			Value: Red200,
		},
		{
			Name:  "Red300",
			Value: Red300,
		},
		{
			Name:  "Red400",
			Value: Red400,
		},
		{
			Name:  "Red500",
			Value: Red500,
		},
		{
			Name:  "Red600",
			Value: Red600,
		},
		{
			Name:  "Red700",
			Value: Red700,
		},
		{
			Name:  "Red800",
			Value: Red800,
		},
		{
			Name:  "Red900",
			Value: Red900,
		},
	}
	OrangeFamily = []ColorPair{
		{
			Name:  "Orange100",
			Value: Orange100,
		},
		{
			Name:  "Orange200",
			Value: Orange200,
		},
		{
			Name:  "Orange300",
			Value: Orange300,
		},
		{
			Name:  "Orange400",
			Value: Orange400,
		},
		{
			Name:  "Orange500",
			Value: Orange500,
		},
		{
			Name:  "Orange600",
			Value: Orange600,
		},
		{
			Name:  "Orange700",
			Value: Orange700,
		},
		{
			Name:  "Orange800",
			Value: Orange800,
		},
		{
			Name:  "Orange900",
			Value: Orange900,
		},
	}
	YellowFamily = []ColorPair{
		{
			Name:  "Yellow100",
			Value: Yellow100,
		},
		{
			Name:  "Yellow200",
			Value: Yellow200,
		},
		{
			Name:  "Yellow300",
			Value: Yellow300,
		},
		{
			Name:  "Yellow400",
			Value: Yellow400,
		},
		{
			Name:  "Yellow500",
			Value: Yellow500,
		},
		{
			Name:  "Yellow600",
			Value: Yellow600,
		},
		{
			Name:  "Yellow700",
			Value: Yellow700,
		},
		{
			Name:  "Yellow800",
			Value: Yellow800,
		},
		{
			Name:  "Yellow900",
			Value: Yellow900,
		},
	}
	GreenFamily = []ColorPair{
		{
			Name:  "Green100",
			Value: Green100,
		},
		{
			Name:  "Green200",
			Value: Green200,
		},
		{
			Name:  "Green300",
			Value: Green300,
		},
		{
			Name:  "Green400",
			Value: Green400,
		},
		{
			Name:  "Green500",
			Value: Green500,
		},
		{
			Name:  "Green600",
			Value: Green600,
		},
		{
			Name:  "Green700",
			Value: Green700,
		},
		{
			Name:  "Green800",
			Value: Green800,
		},
		{
			Name:  "Green900",
			Value: Green900,
		},
	}
	TealFamily = []ColorPair{
		{
			Name:  "Teal100",
			Value: Teal100,
		},
		{
			Name:  "Teal200",
			Value: Teal200,
		},
		{
			Name:  "Teal300",
			Value: Teal300,
		},
		{
			Name:  "Teal400",
			Value: Teal400,
		},
		{
			Name:  "Teal500",
			Value: Teal500,
		},
		{
			Name:  "Teal600",
			Value: Teal600,
		},
		{
			Name:  "Teal700",
			Value: Teal700,
		},
		{
			Name:  "Teal800",
			Value: Teal800,
		},
		{
			Name:  "Teal900",
			Value: Teal900,
		},
	}
	CyanFamily = []ColorPair{
		{
			Name:  "Cyan100",
			Value: Cyan100,
		},
		{
			Name:  "Cyan200",
			Value: Cyan200,
		},
		{
			Name:  "Cyan300",
			Value: Cyan300,
		},
		{
			Name:  "Cyan400",
			Value: Cyan400,
		},
		{
			Name:  "Cyan500",
			Value: Cyan500,
		},
		{
			Name:  "Cyan600",
			Value: Cyan600,
		},
		{
			Name:  "Cyan700",
			Value: Cyan700,
		},
		{
			Name:  "Cyan800",
			Value: Cyan800,
		},
		{
			Name:  "Cyan900",
			Value: Cyan900,
		},
	}
	GrayFamily = []ColorPair{
		{
			Name:  "Gray100",
			Value: Gray100,
		},
		{
			Name:  "Gray200",
			Value: Gray200,
		},
		{
			Name:  "Gray300",
			Value: Gray300,
		},
		{
			Name:  "Gray400",
			Value: Gray400,
		},
		{
			Name:  "Gray500",
			Value: Gray500,
		},
		{
			Name:  "Gray600",
			Value: Gray600,
		},
		{
			Name:  "Gray700",
			Value: Gray700,
		},
		{
			Name:  "Gray800",
			Value: Gray800,
		},
		{
			Name:  "Gray900",
			Value: Gray900,
		},
	}

	ColorsFamilies = []string{
		"Blue",
		"Indigo",
		"Purple",
		"Pink",
		"Red",
		"Orange",
		"Yellow",
		"Green",
		"Teal",
		"Cyan",
		"Gray",
	}

	// All colors, grouped by their family
	ColorsByFamily = map[string][]ColorPair{
		"Blue":   BlueFamily,
		"Indigo": IndigoFamily,
		"Purple": PurpleFamily,
		"Pink":   PinkFamily,
		"Red":    RedFamily,
		"Orange": OrangeFamily,
		"Yellow": YellowFamily,
		"Green":  GreenFamily,
		"Teal":   TealFamily,
		"Cyan":   CyanFamily,
		"Gray":   GrayFamily,
	}

	// All known colors in a map to easily look up their name to value
	AllColors = map[string]lipgloss.Color{
		"blue":       Blue,
		"blue-100":   Blue100,
		"blue-200":   Blue200,
		"blue-300":   Blue300,
		"blue-400":   Blue400,
		"blue-500":   Blue500,
		"blue-600":   Blue600,
		"blue-700":   Blue700,
		"blue-800":   Blue800,
		"blue-900":   Blue900,
		"indigo":     Indigo,
		"indigo-100": Indigo100,
		"indigo-200": Indigo200,
		"indigo-300": Indigo300,
		"indigo-400": Indigo400,
		"indigo-500": Indigo500,
		"indigo-600": Indigo600,
		"indigo-700": Indigo700,
		"indigo-800": Indigo800,
		"indigo-900": Indigo900,
		"purple":     Purple,
		"purple-100": Purple100,
		"purple-200": Purple200,
		"purple-300": Purple300,
		"purple-400": Purple400,
		"purple-500": Purple500,
		"purple-600": Purple600,
		"purple-700": Purple700,
		"purple-800": Purple800,
		"purple-900": Purple900,
		"pink":       Pink,
		"pink-100":   Pink100,
		"pink-200":   Pink200,
		"pink-300":   Pink300,
		"pink-400":   Pink400,
		"pink-500":   Pink500,
		"pink-600":   Pink600,
		"pink-700":   Pink700,
		"pink-800":   Pink800,
		"pink-900":   Pink900,
		"red":        Red,
		"red-100":    Red100,
		"red-200":    Red200,
		"red-300":    Red300,
		"red-400":    Red400,
		"red-500":    Red500,
		"red-600":    Red600,
		"red-700":    Red700,
		"red-800":    Red800,
		"red-900":    Red900,
		"orange":     Orange,
		"orange-100": Orange100,
		"orange-200": Orange200,
		"orange-300": Orange300,
		"orange-400": Orange400,
		"orange-500": Orange500,
		"orange-600": Orange600,
		"orange-700": Orange700,
		"orange-800": Orange800,
		"orange-900": Orange900,
		"yellow":     Yellow,
		"yellow-100": Yellow100,
		"yellow-200": Yellow200,
		"yellow-300": Yellow300,
		"yellow-400": Yellow400,
		"yellow-500": Yellow500,
		"yellow-600": Yellow600,
		"yellow-700": Yellow700,
		"yellow-800": Yellow800,
		"yellow-900": Yellow900,
		"green":      Green,
		"green-100":  Green100,
		"green-200":  Green200,
		"green-300":  Green300,
		"green-400":  Green400,
		"green-500":  Green500,
		"green-600":  Green600,
		"green-700":  Green700,
		"green-800":  Green800,
		"green-900":  Green900,
		"teal":       Teal,
		"teal-100":   Teal100,
		"teal-200":   Teal200,
		"teal-300":   Teal300,
		"teal-400":   Teal400,
		"teal-500":   Teal500,
		"teal-600":   Teal600,
		"teal-700":   Teal700,
		"teal-800":   Teal800,
		"teal-900":   Teal900,
		"cyan":       Cyan,
		"cyan-100":   Cyan100,
		"cyan-200":   Cyan200,
		"cyan-300":   Cyan300,
		"cyan-400":   Cyan400,
		"cyan-500":   Cyan500,
		"cyan-600":   Cyan600,
		"cyan-700":   Cyan700,
		"cyan-800":   Cyan800,
		"cyan-900":   Cyan900,
		"gray":       Gray,
		"gray-100":   Gray100,
		"gray-200":   Gray200,
		"gray-300":   Gray300,
		"gray-400":   Gray400,
		"gray-500":   Gray500,
		"gray-600":   Gray600,
		"gray-700":   Gray700,
		"gray-800":   Gray800,
		"gray-900":   Gray900,
	}
)

func Replace(color string) string {
	if strings.HasPrefix(color, "$") {
		v, ok := AllColors[strings.TrimPrefix(color, "$")]
		if !ok {
			log.Fatalf("Unknown color: %q", color)
		}

		return string(v)
	}

	return color
}
