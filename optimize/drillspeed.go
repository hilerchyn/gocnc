package optimize

import "github.com/joushou/gocnc/vm"
import "github.com/joushou/gocnc/utils"

//
// Ideas for other optimization steps:
//   Move grouping - Group moves based on Z0, Zdepth lifts, to finalize
//      section, instead of constantly moving back and forth
//   Vector-angle removal - Combine moves where the move vector changes
//      less than a certain minimum angle
//

// Detects a previous drill, and uses rapid move to the previous known depth.
// Scans through all Z-descent moves, logs its height, and ensures that any future move
// at that location will use vm.MoveModeRapid to go to the deepest previous known Z-height.
func OptDrillSpeed(machine *vm.Machine) {
	var (
		last       utils.Vector
		npos       []vm.Position = make([]vm.Position, 0)
		drillStack []vm.Position = make([]vm.Position, 0)
	)

	fastDrill := func(pos vm.Position) (vm.Position, vm.Position, bool) {
		var depth float64
		var found bool
		for _, m := range drillStack {
			if m.X == pos.X && m.Y == pos.Y {
				if m.Z < depth {
					depth = m.Z
					found = true
				}
			}
		}

		drillStack = append(drillStack, pos)

		if found {
			if pos.Z >= depth { // We have drilled all of it, so just rapid all the way
				pos.State.MoveMode = vm.MoveModeRapid
				return pos, pos, false
			} else { // Can only rapid some of the way
				p := pos
				p.Z = depth
				p.State.MoveMode = vm.MoveModeRapid
				return p, pos, true
			}
		} else {
			return pos, pos, false
		}
	}

	for _, m := range machine.Positions {
		if m.X == last.X && m.Y == last.Y && m.Z < last.Z && m.State.MoveMode == vm.MoveModeLinear {
			posn, poso, shouldinsert := fastDrill(m)
			if shouldinsert {
				npos = append(npos, posn)
			}
			npos = append(npos, poso)
		} else {
			npos = append(npos, m)
		}
		last = m.Vector()
	}
	machine.Positions = npos
}
