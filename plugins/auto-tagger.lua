-- plugins/auto-tagger.lua
-- Automatically tags tasks based on title keywords
-- Demonstrates event hooks and task updates

local plugin = {
    id = "auto-tagger",
    name = "Auto Tagger",
    description = "Automatically tags tasks based on keywords in the title",
    author = "Kairo",
    version = "1.0.0",

    -- Define tag rules: {keywords, tag}
    rules = {
        { keywords = {"bug", "fix", "error"}, tag = "bug" },
        { keywords = {"feature", "new", "enhancement", "feat"}, tag = "feature" },
        { keywords = {"review", "pr", "merge"}, tag = "review" },
        { keywords = {"doc", "documentation", "readme"}, tag = "docs" },
        { keywords = {"test", "testing", "unit"}, tag = "test" },
    }
}

-- Hook into task creation events
kairo.on("task_create", function(event)
    local task = event.task
    if not task then return end

    local title_lower = string.lower(task.title)
    local matched_tags = {}

    -- Check each rule
    for _, rule in ipairs(plugin.rules) do
        for _, keyword in ipairs(rule.keywords) do
            if string.find(title_lower, keyword, 1, true) then
                table.insert(matched_tags, rule.tag)
                break
            end
        end
    end

    -- Apply matched tags if any
    if #matched_tags > 0 then
        local existing_tags = task.tags or {}
        
        -- Merge tags (avoid duplicates)
        local tag_map = {}
        for _, tag in ipairs(existing_tags) do
            tag_map[tag] = true
        end
        for _, tag in ipairs(matched_tags) do
            if not tag_map[tag] then
                table.insert(existing_tags, tag)
                tag_map[tag] = true
            end
        end

        -- Update task with new tags
        local updated, err = kairo.update_task(task.id, {tags = existing_tags})
        if err then
            kairo.notify("Failed to auto-tag task: " .. err, true)
        end
    end
end)

return plugin
