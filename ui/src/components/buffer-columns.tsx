import { Buffer } from "@/lib/entities";
import { Badge } from "./ui/badge";
import { useRelativeTime } from "@/lib/utils";

export const columns = () => [
  {
    key: "label" as keyof Buffer,
    title: "Label",
    render: (value: string) => <div className="font-medium">{value}</div>,
  },
  {
    key: "component" as keyof Buffer,
    title: "Type",
    render: (value: string) => (
      <Badge
        variant="outline"
        className="bg-purple-50 text-purple-700 dark:bg-purple-900/20 dark:text-purple-400"
      >
        {value}
      </Badge>
    ),
  },
  {
    key: "createdAt" as keyof Buffer,
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
