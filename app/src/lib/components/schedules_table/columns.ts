import type { ColumnDef } from "@tanstack/table-core";
import { createRawSnippet } from "svelte";
import { renderSnippet } from "$lib/components/ui/data-table"; // Adjusted import path

// This type is used to define the shape of our data.
// We are fetching this from the Go backend.
export type Schedule = {
  schedule_id: number;
  name: string;
  cron_expr: string;
  start_date?: string | null; // ISO string
  end_date?: string | null;   // ISO string
  duration_minutes: number;
  timezone?: string | null;
};

// Helper to format date strings or return 'N/A'
const formatDate = (dateString?: string | null): string => {
  if (!dateString) return "N/A";
  try {
    return new Date(dateString).toLocaleDateString();
  } catch {
    // Error object 'e' is not needed here
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
      const snippet = createRawSnippet<[string]>((getExpr) => {
         const val = getExpr();
         return { render: () => `<code>${val}</code>`};
      });
      return renderSnippet(snippet, cronExpr);
    }
  },
  {
    accessorKey: "start_date",
    header: "Start Date",
    cell: ({ row }) => {
      const startDate = formatDate(row.getValue("start_date"));
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
      const endDate = formatDate(row.getValue("end_date"));
       const snippet = createRawSnippet<[string]>((getDate) => {
        const val = getDate();
        return { render: () => `<div>${val}</div>`};
      });
      return renderSnippet(snippet, endDate);
    }
  },
  {
    accessorKey: "duration_minutes",
    header: "Duration (Mins)",
    cell: ({ row }) => {
      const duration = row.getValue("duration_minutes") as number;
      const snippet = createRawSnippet<[number]>((getDuration) => {
        const val = getDuration();
        return { render: () => `<div class="text-right">${val}</div>`}; // Example: right align
      });
      return renderSnippet(snippet, duration);
    }
  },
  {
    accessorKey: "timezone",
    header: "Timezone",
    cell: ({ row }) => {
      const timezone = (row.getValue("timezone") as string | null) || "N/A";
       const snippet = createRawSnippet<[string]>((getTz) => {
        const val = getTz();
        return { render: () => `<div>${val}</div>`};
      });
      return renderSnippet(snippet, timezone);
    }
  },
  // TODO: Consider adding an 'actions' column later if needed
  // {
  //   id: "actions",
  //   cell: ({ row }) => {
  //     // const schedule = row.original;
  //     // return renderComponent(ScheduleActions, { id: schedule.schedule_id });
  //     return "Actions..."; // Placeholder
  //   },
  // },
]; 