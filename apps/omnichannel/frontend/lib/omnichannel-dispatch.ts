import { notificationApi } from '@/lib/api';
import { OmnichannelCampaign, OmnichannelRecipient } from '@/lib/omnichannel-store';

export type CampaignDispatchResult = {
  submitted: number;
  failed: number;
  message: string;
};

export async function dispatchCampaign(
  campaign: OmnichannelCampaign,
  recipients: OmnichannelRecipient[],
): Promise<CampaignDispatchResult> {
  const targets = recipients.filter(
    (recipient) =>
      recipient.status === 'active' &&
      recipient.segment === campaign.segment &&
      recipient.channel === campaign.channel,
  );

  if (targets.length === 0) {
    return {
      submitted: 0,
      failed: 0,
      message: 'No active recipients match this campaign segment/channel.',
    };
  }

  const results = await Promise.allSettled(
    targets.map((recipient) =>
      notificationApi.create({
        channel: campaign.channel,
        recipient: recipient.destination,
        recipient_email: recipient.destination,
        template_id: campaign.template_id,
        subject: campaign.template_id ? undefined : campaign.subject,
        body: campaign.template_id ? undefined : campaign.body,
        metadata: {
          source: 'omnichannel.campaign',
          campaign_id: campaign.id,
          segment: campaign.segment,
        },
      }),
    ),
  );

  const submitted = results.filter((result) => result.status === 'fulfilled').length;
  const failed = results.length - submitted;

  if (submitted === 0) {
    return {
      submitted,
      failed,
      message: 'Campaign dispatch failed for all recipients.',
    };
  }

  if (failed > 0) {
    return {
      submitted,
      failed,
      message: `Dispatched ${submitted} notifications (${failed} failed).`,
    };
  }

  return {
    submitted,
    failed,
    message: `Dispatched ${submitted} notifications.`,
  };
}
