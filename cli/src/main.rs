use anyhow::{bail, Context, Result};
use clap::{Parser, Subcommand, ValueEnum};
use reqwest::blocking::Client;
use reqwest::header::{HeaderMap, HeaderValue, CONTENT_TYPE};
use serde::{Deserialize, Serialize};
use serde_json::json;

#[derive(Parser, Debug)]
#[command(name = "x_cli", about = "CLI for omnichannel notification operations")]
struct Cli {
    #[arg(
        long,
        env = "OMNICHANNEL_API_URL",
        default_value = "http://localhost:8080/api/v1"
    )]
    api_url: String,

    #[arg(long, env = "OMNICHANNEL_API_KEY", default_value = "test-api-key-123")]
    api_key: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Notifications {
        #[command(subcommand)]
        command: NotificationsCommand,
    },
}

#[derive(Subcommand, Debug)]
enum NotificationsCommand {
    Send {
        #[arg(long, value_enum, default_value = "email")]
        channel: Channel,

        #[arg(long)]
        to: String,

        #[arg(long)]
        subject: String,

        #[arg(long)]
        body: String,
    },
    List {
        #[arg(long, default_value_t = 20)]
        limit: usize,
    },
    Status {
        #[arg(long)]
        id: String,
    },
}

#[derive(Debug, Clone, Copy, ValueEnum, Serialize)]
#[serde(rename_all = "lowercase")]
enum Channel {
    Email,
    Sms,
    Push,
    Webhook,
}

#[derive(Debug, Deserialize)]
struct Notification {
    id: String,
    subject: String,
    recipient_email: String,
    status: String,
    metadata: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
struct ListResponse {
    data: Vec<Notification>,
}

#[derive(Debug, Deserialize)]
struct StatusResponse {
    id: String,
    status: String,
    workflow_status: Option<String>,
    failed_at: Option<String>,
    sent_at: Option<String>,
    error_message: Option<String>,
}

fn main() -> Result<()> {
    let cli = Cli::parse();
    let api = OmnichannelApi::new(&cli.api_url, &cli.api_key)?;

    match cli.command {
        Commands::Notifications { command } => match command {
            NotificationsCommand::Send {
                channel,
                to,
                subject,
                body,
            } => {
                let id = api.send_notification(channel, &to, &subject, &body)?;
                println!("created notification: {id}");
            }
            NotificationsCommand::List { limit } => {
                let notifications = api.list_notifications()?;
                for notif in notifications.into_iter().take(limit) {
                    let channel = notif
                        .metadata
                        .as_ref()
                        .and_then(|value| value.get("channel"))
                        .and_then(|value| value.as_str())
                        .unwrap_or("email");

                    println!(
                        "{} | {} | {} | {} | {}",
                        notif.id, notif.status, channel, notif.recipient_email, notif.subject
                    );
                }
            }
            NotificationsCommand::Status { id } => {
                let status = api.notification_status(&id)?;
                println!("id: {}", status.id);
                println!("status: {}", status.status);
                println!(
                    "workflow_status: {}",
                    status
                        .workflow_status
                        .unwrap_or_else(|| "unknown".to_string())
                );
                if let Some(sent_at) = status.sent_at {
                    println!("sent_at: {sent_at}");
                }
                if let Some(failed_at) = status.failed_at {
                    println!("failed_at: {failed_at}");
                }
                if let Some(error_message) = status.error_message {
                    if !error_message.is_empty() {
                        println!("error_message: {error_message}");
                    }
                }
            }
        },
    }

    Ok(())
}

struct OmnichannelApi {
    base_url: String,
    client: Client,
}

impl OmnichannelApi {
    fn new(api_url: &str, api_key: &str) -> Result<Self> {
        let mut headers = HeaderMap::new();
        headers.insert(CONTENT_TYPE, HeaderValue::from_static("application/json"));
        headers.insert(
            "X-API-Key",
            HeaderValue::from_str(api_key).context("invalid API key header")?,
        );

        let client = Client::builder()
            .default_headers(headers)
            .build()
            .context("failed to build HTTP client")?;

        Ok(Self {
            base_url: api_url.trim_end_matches('/').to_string(),
            client,
        })
    }

    fn send_notification(
        &self,
        channel: Channel,
        to: &str,
        subject: &str,
        body: &str,
    ) -> Result<String> {
        let endpoint = format!("{}/notifications", self.base_url);
        let payload = json!({
            "channel": channel,
            "recipient": to,
            "recipient_email": to,
            "subject": subject,
            "body": body,
            "metadata": {
                "source": "x_cli",
                "channel": channel,
                "destination": to,
            }
        });

        let response = self
            .client
            .post(&endpoint)
            .json(&payload)
            .send()
            .with_context(|| format!("failed to call {endpoint}"))?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().unwrap_or_default();
            bail!("send failed: {} {}", status, body);
        }

        let value: serde_json::Value = response
            .json()
            .context("failed to parse create notification response")?;
        let id = value
            .get("id")
            .and_then(|v| v.as_str())
            .context("create response missing id")?;

        Ok(id.to_string())
    }

    fn list_notifications(&self) -> Result<Vec<Notification>> {
        let endpoint = format!("{}/notifications", self.base_url);
        let response = self
            .client
            .get(&endpoint)
            .send()
            .with_context(|| format!("failed to call {endpoint}"))?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().unwrap_or_default();
            bail!("list failed: {} {}", status, body);
        }

        let payload: ListResponse = response
            .json()
            .context("failed to parse notifications list response")?;

        Ok(payload.data)
    }

    fn notification_status(&self, id: &str) -> Result<StatusResponse> {
        let endpoint = format!("{}/notifications/{}/status", self.base_url, id);
        let response = self
            .client
            .get(&endpoint)
            .send()
            .with_context(|| format!("failed to call {endpoint}"))?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().unwrap_or_default();
            bail!("status failed: {} {}", status, body);
        }

        response
            .json::<StatusResponse>()
            .context("failed to parse notification status response")
    }
}
