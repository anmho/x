const CONTROL_PLANE_BASE_URL = (process.env.PLATFORM_CONTROL_PLANE_URL || 'http://127.0.0.1:8091').replace(/\/$/, '');

export class ControlPlaneUnavailableError extends Error {
  constructor(cause?: unknown) {
    super(
      `Control plane is unavailable at ${CONTROL_PLANE_BASE_URL}. Start it with ./platform control-plane serve --addr :8091, or use local fallback mode.`,
      { cause },
    );
    this.name = 'ControlPlaneUnavailableError';
  }
}

export async function callControlPlane(path: string, init?: RequestInit): Promise<Response> {
  try {
    return await fetch(`${CONTROL_PLANE_BASE_URL}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(init?.headers || {}),
      },
      cache: 'no-store',
    });
  } catch (error) {
    throw new ControlPlaneUnavailableError(error);
  }
}
