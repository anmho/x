export function matchesCronField(field: string, value: number): boolean {
  const trimmed = field.trim();
  if (trimmed === '*') return true;
  if (trimmed.startsWith('*/')) {
    const step = Number.parseInt(trimmed.slice(2), 10);
    if (!Number.isFinite(step) || step <= 0) return false;
    return value % step === 0;
  }
  const numeric = Number.parseInt(trimmed, 10);
  if (!Number.isFinite(numeric)) return false;
  return numeric === value;
}

export function isCronDue(
  expression: string,
  now: Date,
  lastRunAt?: string,
): boolean {
  const parts = expression.trim().split(/\s+/);
  if (parts.length !== 5) return false;
  const [minuteField, hourField, dayField, monthField, dayOfWeekField] = parts;

  if (!matchesCronField(minuteField, now.getMinutes())) return false;
  if (!matchesCronField(hourField, now.getHours())) return false;
  if (!matchesCronField(dayField, now.getDate())) return false;
  if (!matchesCronField(monthField, now.getMonth() + 1)) return false;
  if (!matchesCronField(dayOfWeekField, now.getDay())) return false;

  if (!lastRunAt) return true;
  const lastRun = new Date(lastRunAt);
  if (Number.isNaN(lastRun.getTime())) return true;
  return !(
    lastRun.getFullYear() === now.getFullYear() &&
    lastRun.getMonth() === now.getMonth() &&
    lastRun.getDate() === now.getDate() &&
    lastRun.getHours() === now.getHours() &&
    lastRun.getMinutes() === now.getMinutes()
  );
}
