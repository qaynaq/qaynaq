# To The End

Reads the entire stream all the way to EOF and emits it as a single message. Use only when the stream has a clear, bounded end and is small enough to fit in memory.

This scanner takes no configuration.

:::warning
Streams that have no end (like a long-running socket) will accumulate forever. Pick a different scanner for unbounded sources.
:::
