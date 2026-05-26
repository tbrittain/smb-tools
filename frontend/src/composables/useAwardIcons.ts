const AWARD_ICONS: Record<string, string> = {
  'League Champion': '🏆',
  'Conference Champion': '🥈',
}

export function getAwardIcon(originalName: string): string {
  return AWARD_ICONS[originalName] ?? ''
}
