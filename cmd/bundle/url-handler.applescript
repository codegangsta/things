on open location theURL
    -- Extract pipe ID from URL query string
    -- URL format: things-cli://success?pipe=UUID&x-things-id=...
    -- or: things-cli://error?pipe=UUID&errorMessage=...

    set pipeID to my extractQueryParam(theURL, "pipe")

    if pipeID is not "" then
        set pipePath to "/tmp/things-cli-" & pipeID & ".pipe"
        try
            do shell script "echo " & quoted form of theURL & " > " & quoted form of pipePath
        end try
    end if
end open location

on extractQueryParam(theURL, paramName)
    set oldDelims to AppleScript's text item delimiters
    try
        -- Get query string after ?
        set AppleScript's text item delimiters to "?"
        if (count of text items of theURL) < 2 then
            set AppleScript's text item delimiters to oldDelims
            return ""
        end if
        set queryString to text item 2 of theURL

        -- Split by &
        set AppleScript's text item delimiters to "&"
        set params to text items of queryString

        repeat with param in params
            set AppleScript's text item delimiters to "="
            set paramParts to text items of param
            if (count of paramParts) >= 2 and (item 1 of paramParts) is paramName then
                set AppleScript's text item delimiters to oldDelims
                return item 2 of paramParts
            end if
        end repeat

        set AppleScript's text item delimiters to oldDelims
        return ""
    on error
        set AppleScript's text item delimiters to oldDelims
        return ""
    end try
end extractQueryParam
