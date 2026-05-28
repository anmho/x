import { NextRequest, NextResponse } from 'next/server';

type EmulatorMessage = {
  id: string;
  channel: string;
  destination: string;
  title: string;
  body: string;
  sent_at: string;
};

const globalStore = globalThis as typeof globalThis & {
  __omnichannelEmulatorMessages?: EmulatorMessage[];
};

function getStore() {
  if (!globalStore.__omnichannelEmulatorMessages) {
    globalStore.__omnichannelEmulatorMessages = [];
  }
  return globalStore.__omnichannelEmulatorMessages;
}

export async function GET() {
  return NextResponse.json({
    data: getStore(),
  });
}

export async function POST(request: NextRequest) {
  const payload = await request.json();
  const message: EmulatorMessage = {
    id: String(Date.now()),
    channel: String(payload.channel || 'app'),
    destination: String(payload.destination || ''),
    title: String(payload.title || ''),
    body: String(payload.body || ''),
    sent_at: String(payload.sent_at || new Date().toISOString()),
  };

  const store = getStore();
  store.unshift(message);
  if (store.length > 200) {
    store.splice(200);
  }

  return NextResponse.json({ ok: true, message }, { status: 201 });
}

export async function DELETE() {
  const store = getStore();
  store.splice(0, store.length);
  return NextResponse.json({ ok: true });
}
