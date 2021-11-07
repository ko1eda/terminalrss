## TODO
---
+ Add author field [x]
+ Add Source as own type to support both http sources and file types sources [x]
+ Move Load http call to own method that Load method calls only if the source is an http source
+ Remove collection type from processor and have it return Feed type
+ Tests
+ Support FOR RSS V1 
+ TUI (might need viper as well -- look into this)
+ Create or add thread safe logger 
+ Normalize More fields from rss 
+ Validate StringsToSources method better 


## Potential TODO
---
+ Make Add sources use an io.Reader we can do this under the hood so the add sources method of the client still uses an easy syntax but this will also support testing with golden files without parsing the rss from the files first 