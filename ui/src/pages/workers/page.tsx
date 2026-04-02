import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { InfoIcon } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { Worker } from "@/lib/entities";
import { DataTable } from "@/components/data-table";
import { fetchWorkers } from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

export default function WorkersPage() {
  const [workers, setWorkers] = useState<Worker[]>([] as Worker[]);

  useEffect(() => {
    async function loadWorkers() {
      try {
        const data = await fetchWorkers();
        setWorkers(data);
      } catch (error) {
        console.error("Error fetching workers data:", error);
      }
    }

    loadWorkers();
  }, []);

  const columns = [
    { key: "id" as keyof Worker, title: "ID" },
    { key: "address" as keyof Worker, title: "Address" },
    {
      key: "status" as keyof Worker,
      title: "Status",
      render: (value: string) => {
        const colorMap: Record<string, string> = {
          active:
            "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
          inactive:
            "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300",
        };
        return (
          <Badge
            className={colorMap[value.toLowerCase()] || ""}
            variant="outline"
          >
            {value}
          </Badge>
        );
      },
    },
    {
      key: "activeFlows" as keyof Worker,
      title: "Active Flows",
    },
    {
      key: "lastHeartbeat" as keyof Worker,
      title: "Last heartbeat",
      render: (value: string) => {
        const RelativeTime = () => {
          const time = useRelativeTime(value);
          return <>{time}</>;
        };
        return <RelativeTime />;
      },
    },
  ];

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Workers</h1>
          <p className="text-muted-foreground">
            Workers automatically join the system and cannot be manually created
            or modified
          </p>
        </div>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="flex items-center text-muted-foreground cursor-help">
                <InfoIcon className="h-5 w-5 mr-1" />
                <span>About Workers</span>
              </div>
            </TooltipTrigger>
            <TooltipContent className="max-w-sm">
              <p>
                Workers are processing nodes that automatically register
                themselves with the system. They report their status via
                heartbeats and can run multiple flows simultaneously.
              </p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Active Workers</CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable data={workers} columns={columns} />
        </CardContent>
      </Card>
    </div>
  );
}
