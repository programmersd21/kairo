return {
  id = "sample",
  commands = {
    {
      id = "capture",
      title = "Capture: quick task",
      hint = "Creates a task tagged #inbox",
      run = function()
        kairo.create_task("Captured task", "Created by sample plugin.", {"inbox"}, 1, "todo")
      end
    }
  },
  views = {
    {
      id = "focus",
      title = "Focus (Doing • P2+)",
      filter = {
        statuses = {"doing"},
        min_priority = 2
      }
    }
  }
}

