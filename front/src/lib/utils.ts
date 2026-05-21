import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatAge(seconds: number) {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "-";
  }

  const units = [
    { label: "d", value: 86400 },
    { label: "h", value: 3600 },
    { label: "m", value: 60 }
  ];

  for (const unit of units) {
    if (seconds >= unit.value) {
      return `${Math.floor(seconds / unit.value)}${unit.label}`;
    }
  }

  return `${Math.floor(seconds)}s`;
}
