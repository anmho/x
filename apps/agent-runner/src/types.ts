export interface RunResource {
  uri: string;
  title?: string;
  mimeType?: string;
  text?: string;
}

export interface RunEnvelope {
  id: string;
  message: string;
  status: string;
  resources?: RunResource[];
}
