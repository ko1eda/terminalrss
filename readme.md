## TODO

+ Add author field [x]
+ Add Source as own type to support both http sources and file types sources [x]
+ Add file reading support [x]
+ Remove collection type from processor and have it return Feed type [x]
+ Tests - in progress!
+ Support FOR RSS V1 
+ Break source management into its own component to remove excess source related structs from client [x]
    - DO this with feeder and other types of components the client relies on. 
    - Replace concrete types with interfaces (Sourcer Interface in place of passing SourceMap struct -- only if this improves the code)
+ TUI (might need viper as well -- look into this)
+ Create or add thread safe logger 
+ Normalize More fields from rss [x]
+ Validate MapToSource method better 
+ Add Client config setting for where to load files from 
