export const AUTH_COOKIE_NAME = 'console_auth';

export function getConsolePassword(): string {
  return process.env.CONSOLE_PASSWORD || 'changeme';
}

export function isConsoleAuthDisabled(): boolean {
  const explicit = (process.env.CONSOLE_AUTH_DISABLED || '').trim().toLowerCase();
  if (explicit === '1' || explicit === 'true' || explicit === 'yes') {
    return true;
  }
  if (explicit === '0' || explicit === 'false' || explicit === 'no') {
    return false;
  }
  return process.env.NODE_ENV !== 'production';
}
