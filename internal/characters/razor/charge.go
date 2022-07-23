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
	spinHitmark   = 7
	spinFrames    = 23
	finishHitmark = 28
	finishFrames  = 51
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

	//Charge Attacks are disabled during Razor's burst
	if c.StatusIsActive(burstBuffKey) {
		return action.ActionInfo{
			Frames:          frames.NewAbilFunc(frames.InitAbilSlice(0)),
			AnimationLength: 0,
			CanQueueAfter:   0,
			State:           c.Core.Player.CurrentState(),
		}
	}

	//TODO: Stamina + Hitlag
	var remainingDuration int = p["duration"] - startFrames
	var spinCount int = 0
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
			break
		}
	}

	//add final spin frames
	totalSpinFrames += 3

	if p["no_finish"] == 0 {
		ai.Mult = charge[1][c.TalentLvlAttack()]
		ai.Abil = "Charge Attack Finishing Slash"
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy),
			totalSpinFrames+finishHitmark,
			totalSpinFrames+finishHitmark,
		)
	}

	var cancelFrame int = totalSpinFrames + finishHitmark
	if p["no_finish"] > 0 {
		cancelFrame = totalSpinFrames - spinFrames - 3 + spinHitmark

		//Set the finisher in a task in case the next action is unable to actually cancel the spin, so the attack still occurs
		c.QueueCharTask(func() {
			if c.Core.Player.CharIsActive(c.Base.Key) && c.Core.Player.CurrentState() == action.ChargeAttackState {
				ai.Mult = charge[1][c.TalentLvlAttack()]
				ai.Abil = "Charge Attack Finishing Slash"
				c.Core.QueueAttack(
					ai,
					combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy),
					0,
					0,
				)
			}
		}, totalSpinFrames+finishHitmark)
	}

	totalFrames = frames.InitAbilSlice(totalSpinFrames + finishFrames)
	totalFrames[action.ActionDash] = cancelFrame
	totalFrames[action.ActionJump] = cancelFrame
	totalFrames[action.ActionSwap] = cancelFrame

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(totalFrames),
		AnimationLength: totalFrames[action.InvalidAction],
		CanQueueAfter:   cancelFrame,
		State:           action.ChargeAttackState,
	}
}
