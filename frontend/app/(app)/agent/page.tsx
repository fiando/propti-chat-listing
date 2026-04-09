import { redirect } from 'next/navigation';
import { getServerAuthProfile } from '@/lib/server-profile';
import { AgentWorkspaceClient } from '@/components/agent/AgentWorkspaceClient';

export default async function AgentPage() {
  const { isAuthenticated } = await getServerAuthProfile();

  if (!isAuthenticated) {
    redirect(`/login?callbackUrl=${encodeURIComponent('/agent')}`);
  }

  return <AgentWorkspaceClient />;
}
