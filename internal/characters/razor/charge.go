package razor

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var spinFrames []int
var finishFrames []int

const (
	startFrames   = 67
	spinHitmark   = 12
	finishHitmark = 30
)

const maxSpinTime = 300

func init() {
	spinFrames = frames.InitAbilSlice(26)
	finishFrames = frames.InitAbilSlice(48)
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

		totalSpinFrames += spinFrames[action.InvalidAction]
		remainingDuration -= spinFrames[action.InvalidAction]

		//Stop spinning if we're past maximum allowed duration (5 seconds), or player has let go
		if remainingDuration <= 0 || totalSpinFrames > maxSpinTime {
			autoEnd = true
			break
		}

		//A spin's frames are cut short when followed by another spin
		totalSpinFrames -= 3
		remainingDuration += 3
	}

	if p["no_finish"] > 0 && !autoEnd {
		return action.ActionInfo{
			Frames:          frames.NewAbilFunc(finishFrames),
			AnimationLength: totalSpinFrames,
			CanQueueAfter:   totalSpinFrames - spinFrames[action.InvalidAction] + spinHitmark,
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

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(finishFrames),
		AnimationLength: totalSpinFrames + finishFrames[action.InvalidAction],
		CanQueueAfter:   totalSpinFrames + finishHitmark,
		State:           action.ChargeAttackState,
	}
}
