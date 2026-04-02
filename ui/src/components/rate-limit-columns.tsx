import { RateLimit } from "@/lib/entities";
import { Badge } from "./ui/badge";
import { useRelativeTime } from "@/lib/utils";

export const columns = () => [
  {
    key: "label" as keyof RateLimit,
    title: "Label",
    render: (value: string) => <div className="font-medium">{value}</div>,
  },
  {
    key: "component" as keyof RateLimit,
    title: "Type",
    render: (value: string) => (
      <Badge
        variant="outline"
        className="bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400"
      >
        {value}
      </Badge>
    ),
  },
  {
    key: "createdAt" as keyof RateLimit,
    title: "Last versioned",
    render: (value: string) => {
      const RelativeTime = () => {
        const time = useRelativeTime(value);
        return <>{time}</>;
      };
      return <RelativeTime />;
    },
  },
];
