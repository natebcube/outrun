package muxhandlers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/fluofoxxo/outrun/config"
	"github.com/fluofoxxo/outrun/consts"
	"github.com/fluofoxxo/outrun/db"
	"github.com/fluofoxxo/outrun/emess"
	"github.com/fluofoxxo/outrun/enums"
	"github.com/fluofoxxo/outrun/helper"
	"github.com/fluofoxxo/outrun/netobj"
	"github.com/fluofoxxo/outrun/requests"
	"github.com/fluofoxxo/outrun/responses"
	"github.com/fluofoxxo/outrun/status"
)

func GetWheelOptions(helper *helper.Helper) {
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	//player.LastWheelOptions = netobj.DefaultWheelOptions(player.PlayerState) // generate new wheel for 'reroll' mechanic
	endPeriod := player.RouletteInfo.RoulettePeriodEnd
	if time.Now().Unix() < endPeriod {
		player.RouletteInfo.RouletteCountInPeriod = 0
		player.RouletteInfo.GotJackpotThisPeriod = false
	}
	response := responses.WheelOptions(baseInfo, player.LastWheelOptions)
	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func CommitWheelSpin(helper *helper.Helper) {
	// TODO: Should this be player agnostic? Make executive decision another time
	recv := helper.GetGameRequest()
	var request requests.CommitWheelSpinRequest
	err := json.Unmarshal(recv, &request)
	if err != nil {
		helper.Err("Error unmarshalling", err)
		return
	}
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}
	if config.CFile.DebugPrints {
		helper.Out(strconv.Itoa(int(request.Count)))
	}

	responseStatus := status.OK
	if player.PlayerState.NumRouletteTicket > 0 { // if we have tickets left
		wonItem := player.LastWheelOptions.Items[player.LastWheelOptions.ItemWon]
		itemExists := player.IndexOfItem(wonItem) != -1
		if itemExists {
			amountOfItemWon := player.LastWheelOptions.Item[player.LastWheelOptions.ItemWon]
			if config.CFile.DebugPrints {
				helper.Out(wonItem)
				helper.Out(strconv.Itoa(int(amountOfItemWon)))
			}
			itemIndex := player.IndexOfItem(wonItem)
			if config.CFile.DebugPrints {
				helper.Out(strconv.Itoa(int(player.PlayerState.Items[itemIndex].Amount)))
			}
			player.PlayerState.Items[itemIndex].Amount += amountOfItemWon
			if config.CFile.DebugPrints {
				helper.Out(strconv.Itoa(int(player.PlayerState.Items[itemIndex].Amount)))
			}
		} else {
			if wonItem == strconv.Itoa(enums.IDTypeItemRouletteWin) {
				// Jackpot
				player.PlayerState.NumRings += player.LastWheelOptions.NumJackpotRing
			} else if wonItem == strconv.Itoa(enums.IDTypeRedRing) {
				// Red rings
				player.PlayerState.NumRedRings += player.LastWheelOptions.Item[player.LastWheelOptions.ItemWon]
			} else {
				helper.Warn("item '" + wonItem + "' not found")
			}
		}

		endPeriod := player.RouletteInfo.RoulettePeriodEnd
		if time.Now().Unix() > endPeriod {
			player.RouletteInfo = netobj.DefaultRouletteInfo() // Effectively reset everything, set new end time
		}

		// generate NEXT! wheel
		player.RouletteInfo.RouletteCountInPeriod++ // we've spun an additional time
		if player.RouletteInfo.RouletteCountInPeriod > consts.RouletteFreeSpins {
			// we've run out of free spins for the period
			player.PlayerState.NumRouletteTicket--
		}
		numRouletteTicket := player.PlayerState.NumRouletteTicket
		rouletteCount := player.RouletteInfo.RouletteCountInPeriod                             // get amount of times we've spun the wheel today
		player.LastWheelOptions = netobj.DefaultWheelOptions(numRouletteTicket, rouletteCount) // create wheel
		if player.RouletteInfo.GotJackpotThisPeriod {
			player.LastWheelOptions.NumJackpotRing = 1
		}
		if wonItem == strconv.Itoa(enums.IDTypeItemRouletteWin) {
			player.RouletteInfo.GotJackpotThisPeriod = true
		}
	} else {
		// do not modify the wheel, set error status
		responseStatus = status.RouletteUseLimit
	}

	baseInfo := helper.BaseInfo(emess.OK, responseStatus)
	response := responses.WheelSpin(baseInfo, player.PlayerState, player.CharacterState, player.ChaoState, player.LastWheelOptions)

	err = db.SavePlayer(player)
	if err != nil {
		helper.InternalErr("Error saving player", err)
		return
	}

	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}
