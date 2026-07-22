import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { main } from '../../wailsjs/go/models'
import SaveFilePicker from './SaveFilePicker.vue'

function candidate(path: string, leagueGUID: string): main.SaveFileCandidateDTO {
  return {
    path,
    gameVersion: 'smb4',
    leagueName: 'Test League',
    numSeasons: 3,
    mode: 'franchise',
    isFranchise: true,
    playerTeamName: 'Test Team',
    leagueGUID,
  } as main.SaveFileCandidateDTO
}

describe('SaveFilePicker', () => {
  it('explains and prevents selection of disabled candidates', async () => {
    const disabled = candidate('C:\\SMB4\\league-connected.sav', 'connected-guid')
    const enabled = candidate('C:\\SMB4\\league-fork.sav', 'fork-guid')
    const reason = 'Already connected. Fork requires a different SMB4 league/save.'
    const wrapper = mount(SaveFilePicker, {
      props: {
        candidates: [disabled, enabled],
        disabledCandidateReasons: { [disabled.path]: reason },
      },
    })

    const cards = wrapper.findAll('.candidate-card')
    expect(cards[0].attributes('aria-disabled')).toBe('true')
    expect(cards[0].text()).toContain(reason)
    expect(cards[0].find('input').attributes('disabled')).toBeDefined()

    await cards[0].trigger('click')
    await cards[0].find('input').trigger('change')
    expect(wrapper.emitted('change')).toBeUndefined()

    await cards[1].trigger('click')
    expect(wrapper.emitted('change')).toEqual([[enabled.path, enabled.leagueGUID, enabled]])
  })
})
