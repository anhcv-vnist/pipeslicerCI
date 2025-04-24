import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export function formatDate(date: string | Date): string {
  if (!date) return 'N/A'
  
  try {
    // Parse the date string into a Date object
    const d = new Date(date)
    
    // Check if the date is valid
    if (isNaN(d.getTime())) {
      console.warn('Invalid date format:', date)
      return 'Invalid date'
    }
    
    // Format options for a more readable date
    const options: Intl.DateTimeFormatOptions = {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZoneName: 'short'
    }
    
    // Use the browser's locale for formatting
    return d.toLocaleString(undefined, options)
  } catch (error) {
    console.error('Error formatting date:', error, date)
    return 'Invalid date'
  }
}
