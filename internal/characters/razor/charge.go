package razor

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var totalFrames []int

const (
	startFrames   = 67
	spinHitmark   = 12
	spinFrames    = 23
	finishHitmark = 30
	finishFrames  = 48
	maxSpinTime   = 300
)

func init() {
}

func (c *char) ChargeAttack(p map[string]int) action.ActionInfo {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		AttackTag:  combat.AttackTagExtra,
		ICDTag:     combat.ICDTagNormalAttack,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Physical,
		Durability: 25,
	}

	//TODO: Stamina stuff

	var remainingDuration int = p["duration"] - startFrames
	var spinCount int = 0
	var autoEnd bool = remainingDuration <= 0

	var totalSpinFrames int = startFrames

	for remainingDuration > 0 {
		spinCount++
		ai.Mult = charge[0][c.TalentLvlAttack()]
		ai.Abil = fmt.Sprintf("Charge Attack Spin %v", spinCount)
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(c.Core.Combat.Player(), 1, false, combat.TargettableEnemy),
			totalSpinFrames+spinHitmark,
			totalSpinFrames+spinHitmark,
		)

		totalSpinFrames += spinFrames
		remainingDuration -= spinFrames

		//Stop spinning if we're past maximum allowed duration (5 seconds)
		if totalSpinFrames > maxSpinTime {
			totalSpinFrames += 3
			autoEnd = true
			break
		}
	}

	if p["no_finish"] > 0 && !autoEnd {
		totalFrames = frames.InitAbilSlice(totalSpinFrames)
		totalFrames[action.ActionDash] = totalSpinFrames - spinFrames - 3 + spinHitmark
		totalFrames[action.ActionJump] = totalSpinFrames - spinFrames - 3 + spinHitmark

		return action.ActionInfo{
			Frames:          frames.NewAbilFunc(totalFrames),
			AnimationLength: totalFrames[action.InvalidAction],
			CanQueueAfter:   totalSpinFrames - spinFrames - 3 + spinHitmark,
			State:           action.ChargeAttackState,
		}
	}

	ai.Mult = charge[1][c.TalentLvlAttack()]
	ai.Abil = "Charge Attack Finishing Slash"
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy),
		totalSpinFrames+finishHitmark,
		totalSpinFrames+finishHitmark,
	)

	totalFrames = frames.InitAbilSlice(totalSpinFrames + finishFrames)
	totalFrames[action.ActionDash] = totalSpinFrames + finishHitmark
	totalFrames[action.ActionJump] = totalSpinFrames + finishHitmark

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(totalFrames),
		AnimationLength: totalFrames[action.InvalidAction],
		CanQueueAfter:   totalSpinFrames + finishHitmark,
		State:           action.ChargeAttackState,
	}
}
