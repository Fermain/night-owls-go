import type { ColumnDef } from "@tanstack/table-core";
import { createRawSnippet } from "svelte";
import { renderSnippet, renderComponent } from "$lib/components/ui/data-table"; // Adjusted import path
import { CronVisualizer } from "$lib/components/cron_visualizer"; // Import the new visualizer
// import { Button } from "$lib/components/ui/button"; // Button import removed as not directly used
import ScheduleActions from "./ScheduleActions.svelte"; // Import the new component

// Define types for how Go's sql.NullString and sql.NullTime are serialized to JSON
export type SQLNullString = {
  String: string;
  Valid: boolean;
};

export type SQLNullTime = {
  Time: string; // Typically an ISO 8601 string from Go's JSON marshaller for time.Time
  Valid: boolean;
};

// This type is used to define the shape of our data.
// We are fetching this from the Go backend.
export type Schedule = {
  schedule_id: number;
  name: string;
  cron_expr: string;
  start_date?: SQLNullTime | null; 
  end_date?: SQLNullTime | null;   
  timezone?: SQLNullString | null; 
};

// Helper to format date strings from SQLNullTime or return 'N/A'
const formatDate = (sqlTime?: SQLNullTime | null): string => {
  if (!sqlTime || !sqlTime.Valid || !sqlTime.Time) return "N/A";
  try {
    // The date from Go might be a full timestamp, extract date part for display.
    return new Date(sqlTime.Time).toLocaleDateString(undefined, { year: 'numeric', month: '2-digit', day: '2-digit' });
  } catch {
    return "Invalid Date";
  }
};

export const columns: ColumnDef<Schedule>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => {
      const name = row.getValue("name") as string;
      // Using createRawSnippet for simple text display; can be enhanced
      const nameSnippet = createRawSnippet<[string]>((getName) => {
        const val = getName();
        return {
          render: () => `<div>${val}</div>`
        };
      });
      return renderSnippet(nameSnippet, name);
    }
  },
  {
    accessorKey: "cron_expr",
    header: "Cron Expression",
    cell: ({ row }) => {
      const cronExpr = row.getValue("cron_expr") as string;
      // Display raw CRON expression string
      const cronSnippet = createRawSnippet<[string]>((getExpr) => {
        const val = getExpr();
        return {
          render: () => `<pre>${val}</pre>` 
        };
      });
      return renderSnippet(cronSnippet, cronExpr);
    }
  },
  {
    id: "cron_visualization",
    header: "Schedule Visualization",
    cell: ({ row }) => {
      const cronExpr = row.original.cron_expr;
      return renderComponent(CronVisualizer, { cronExpr: cronExpr });
    },
    enableSorting: false,
  },
  {
    accessorKey: "start_date",
    header: "Start Date",
    cell: ({ row }) => {
      const startDate = formatDate(row.getValue("start_date") as SQLNullTime | null);
      const snippet = createRawSnippet<[string]>((getDate) => {
        const val = getDate();
        return { render: () => `<div>${val}</div>`};
      });
      return renderSnippet(snippet, startDate);
    }
  },
  {
    accessorKey: "end_date",
    header: "End Date",
    cell: ({ row }) => {
      const endDate = formatDate(row.getValue("end_date") as SQLNullTime | null);
       const snippet = createRawSnippet<[string]>((getDate) => {
        const val = getDate();
        return { render: () => `<div>${val}</div>`};
      });
      return renderSnippet(snippet, endDate);
    }
  },
  {
    id: "actions",
    header: "Actions",
    cell: ({ row }) => {
      const schedule = row.original;
      return renderComponent(ScheduleActions, { scheduleId: schedule.schedule_id });
    },
    enableSorting: false,
    enableHiding: false,
  },
]; 