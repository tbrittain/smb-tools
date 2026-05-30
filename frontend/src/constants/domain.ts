// Domain constants for Super Mega Baseball player attributes.

export const BATTING_POSITIONS = ['C', '1B', '2B', '3B', 'SS', 'LF', 'CF', 'RF', 'DH'] as const
export const PITCHING_ROLES = ['SP', 'RP', 'SP/RP', 'CL'] as const
export const BAT_HANDS = ['L', 'R', 'S'] as const
export const THROW_HANDS = ['L', 'R'] as const
export const CHEMISTRY_TYPES = ['Competitive', 'Spirited', 'Disciplined', 'Scholarly', 'Crafty'] as const

export type BattingPosition = (typeof BATTING_POSITIONS)[number]
export type PitchingRole = (typeof PITCHING_ROLES)[number]
export type BatHand = (typeof BAT_HANDS)[number]
export type ThrowHand = (typeof THROW_HANDS)[number]
export type ChemistryType = (typeof CHEMISTRY_TYPES)[number]
