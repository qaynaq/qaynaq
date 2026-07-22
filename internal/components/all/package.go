package all

import (
	_ "github.com/qaynaq/qaynaq/internal/components/ai_gateway"
	_ "github.com/qaynaq/qaynaq/internal/components/bloblang"
	_ "github.com/qaynaq/qaynaq/internal/components/cdc_mysql"
	_ "github.com/qaynaq/qaynaq/internal/components/coordinator_ratelimit"
	_ "github.com/qaynaq/qaynaq/internal/components/google_calendar"
	_ "github.com/qaynaq/qaynaq/internal/components/google_drive"
	_ "github.com/qaynaq/qaynaq/internal/components/google_sheets"
	_ "github.com/qaynaq/qaynaq/internal/components/mcp_call"
	_ "github.com/qaynaq/qaynaq/internal/components/rag_chunker"
	_ "github.com/qaynaq/qaynaq/internal/components/shopify"

	_ "github.com/warpstreamlabs/bento/public/components/amqp09"
	_ "github.com/warpstreamlabs/bento/public/components/confluent"
	_ "github.com/warpstreamlabs/bento/public/components/dgraph"
	_ "github.com/warpstreamlabs/bento/public/components/huggingface"
	_ "github.com/warpstreamlabs/bento/public/components/io"
	_ "github.com/warpstreamlabs/bento/public/components/kafka"
	_ "github.com/warpstreamlabs/bento/public/components/memcached"
	_ "github.com/warpstreamlabs/bento/public/components/pure"
	_ "github.com/warpstreamlabs/bento/public/components/pure/extended"
	_ "github.com/warpstreamlabs/bento/public/components/redis"
	_ "github.com/warpstreamlabs/bento/public/components/sql"
)
