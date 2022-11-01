package consts

import "github.com/fluofoxxo/outrun/enums"

type PrizeInfo struct {
	AppearanceChance float64 // % chance for it to be chosen to be in wheel by the server
	Type             int64   // 0 for Chao, 1 for Character
}

// A 'load' as depicted below is the chance for the server to pick
// the associated item, where chosen is if randFloat(0, 100) < load.
// IMPORTANT: This load is exclusive to the rarity of the Chao that
// is being chosen by the server.

var RandomChaoWheelCharacterPrizes = map[string]float64{
	// characterID: load
	// Hopefully this should sum up to 100 just for
	// simplicity, but it shouldn't be a requirement.
	enums.CTStrSonic:        10.0,
	enums.CTStrTails:        10.0,
	enums.CTStrKnuckles:     10.0,
}

var RandomChaoWheelChaoPrizes = map[string]float64{
	// TODO: Balance these
	enums.ChaoIDStrHeroChao:             5.0,
	enums.ChaoIDStrGoldChao:             5.0,
	enums.ChaoIDStrDarkChao:             5.0,
	enums.ChaoIDStrJewelChao:            4.5,
	enums.ChaoIDStrNormalChao:           4.5,
	enums.ChaoIDStrOmochao:              4.5,
}
