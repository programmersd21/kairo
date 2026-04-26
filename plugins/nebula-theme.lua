local plugin = {
    id = "nebula-theme",
    name = "Nebula Theme Pack",
    description = "A premium deep-space theme with vibrant nebula accents.",
    author = "Kairo Team",
    version = "1.0.0",
}

-- Define custom themes
plugin.themes = {
    {
        name = "nebula",
        is_light = false,
        bg = "#0B0E14",      -- Deep Space
        fg = "#E1E9F0",      -- Star White
        muted = "#4B5263",   -- Space Dust
        border = "#1C2028",  -- Structural
        accent = "#BD93F9",  -- Nebula Purple
        good = "#50FA7B",    -- Gas Green
        warn = "#FFB86C",    -- Star Orange
        bad = "#FF5555",     -- Nova Red
        overlay = "#151921", -- Internal UI
    },
    {
        name = "nebula_alt",
        is_light = false,
        bg = "#020617",      -- Void Black
        fg = "#F8FAFC",      -- Pure White
        muted = "#334155",   -- Muted Slate
        border = "#1E293B",  -- Slate Border
        accent = "#F472B6",  -- Nebula Pink
        good = "#34D399",    -- Emerald
        warn = "#FBBF24",    -- Amber
        bad = "#F87171",     -- Rose
        overlay = "#0F172A", -- Deep Slate
    }
}

-- Optional: Add a command to notify about the theme
plugin.commands = {
    {
        id = "about-nebula",
        title = "About Nebula Theme",
        hint = "Show theme info",
        run = function()
            kairo.notify("Nebula theme pack loaded! Open the Theme Menu (t) to apply.")
        end
    }
}

return plugin
