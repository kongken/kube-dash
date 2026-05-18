import * as React from "react";
import { cn } from "@/lib/utils";

type BadgeTone = "default" | "success" | "warning" | "muted" | "destructive";

const toneClasses: Record<BadgeTone, string> = {
  default: "bg-primary text-primary-foreground",
  success: "bg-emerald-100 text-emerald-800",
  warning: "bg-amber-100 text-amber-800",
  muted: "bg-muted text-muted-foreground",
  destructive: "bg-destructive text-destructive-foreground"
};

export function Badge({
  className,
  tone = "default",
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & { tone?: BadgeTone }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium",
        toneClasses[tone],
        className
      )}
      {...props}
    />
  );
}
