package app

import "github.com/programmersd21/kairo/internal/core"

type errMsg struct{ Err error }

type tasksLoadedMsg struct{ Tasks []core.Task }
type tagsLoadedMsg struct{ Tags []string }
type allTasksLoadedMsg struct{ Tasks []core.Task }

type taskCreatedMsg struct{ Task core.Task }
type taskUpdatedMsg struct{ Task core.Task }
type taskDeletedMsg struct{ ID string }

type openTaskMsg struct{ Task core.Task }
type openEditMsg struct{ Task core.Task }

type pluginChangedMsg struct{}

type syncDoneMsg struct{ Err error }

type strikeAnimationTickMsg struct{ TaskID string }
