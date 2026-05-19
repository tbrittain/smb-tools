package service

import "smb-tools/internal/models"

// ComputeBattingRates fills the rate fields on b from its counting stats.
// Rate fields are set to nil when their denominator is zero.
// b must not be nil.
func ComputeBattingRates(b *models.CareerBattingStats) {
	if b.AtBats > 0 {
		ba := float64(b.Hits) / float64(b.AtBats)
		b.BA = &ba

		tb := float64(b.Hits-b.Doubles-b.Triples-b.HomeRuns) +
			float64(b.Doubles)*2 +
			float64(b.Triples)*3 +
			float64(b.HomeRuns)*4
		slg := tb / float64(b.AtBats)
		b.SLG = &slg

		iso := slg - ba
		b.ISO = &iso
	}

	obpDenom := b.AtBats + b.Walks + b.HitByPitch + b.SacFlies
	if obpDenom > 0 {
		obp := float64(b.Hits+b.Walks+b.HitByPitch) / float64(obpDenom)
		b.OBP = &obp

		if b.BA != nil && b.SLG != nil {
			ops := obp + *b.SLG
			b.OPS = &ops
		}
	}

	babipDenom := b.AtBats - b.Strikeouts - b.HomeRuns + b.SacFlies
	if babipDenom > 0 {
		babip := float64(b.Hits-b.HomeRuns) / float64(babipDenom)
		b.BABIP = &babip
	}

	pa := b.AtBats + b.Walks + b.HitByPitch + b.SacHits + b.SacFlies
	if pa > 0 {
		kp := float64(b.Strikeouts) / float64(pa)
		b.KPct = &kp
		bbp := float64(b.Walks) / float64(pa)
		b.BBPct = &bbp
	}

	if b.HomeRuns > 0 {
		abhr := float64(b.AtBats) / float64(b.HomeRuns)
		b.ABPerHR = &abhr
	}
}

// ComputePitchingRates fills the rate fields on p from its counting stats.
// Rate fields are set to nil when their denominator is zero.
// p must not be nil.
func ComputePitchingRates(p *models.CareerPitchingStats) {
	if p.OutsPitched > 0 {
		era := float64(p.EarnedRuns) * 27.0 / float64(p.OutsPitched)
		p.ERA = &era

		whip := float64(p.Walks+p.HitsAllowed) * 3.0 / float64(p.OutsPitched)
		p.WHIP = &whip

		k9 := float64(p.Strikeouts) * 27.0 / float64(p.OutsPitched)
		p.K9 = &k9

		bb9 := float64(p.Walks) * 27.0 / float64(p.OutsPitched)
		p.BB9 = &bb9

		h9 := float64(p.HitsAllowed) * 27.0 / float64(p.OutsPitched)
		p.H9 = &h9

		hr9 := float64(p.HomeRunsAllowed) * 27.0 / float64(p.OutsPitched)
		p.HR9 = &hr9

		ppip := float64(p.TotalPitches) * 3.0 / float64(p.OutsPitched)
		p.PPerIP = &ppip
	}

	if p.Walks > 0 {
		kbb := float64(p.Strikeouts) / float64(p.Walks)
		p.KPerBB = &kbb
	}

	if p.BattersFaced > 0 {
		kpct := float64(p.Strikeouts) / float64(p.BattersFaced)
		p.KPct = &kpct
	}

	wl := p.Wins + p.Losses
	if wl > 0 {
		wp := float64(p.Wins) / float64(wl)
		p.WinPct = &wp
	}
}
