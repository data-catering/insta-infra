import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import VersionDisplay from '../VersionDisplay'

// Mock the Wails functions
vi.mock('../../../wailsjs/go/main/App', () => ({
  GetInstaInfraVersion: vi.fn(),
}))

import { GetInstaInfraVersion } from '../../../wailsjs/go/main/App'

describe('VersionDisplay', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('renders loading state initially', () => {
    GetInstaInfraVersion.mockImplementation(() => new Promise(() => {})) // Never resolves
    
    render(<VersionDisplay />)
    
    expect(screen.getByText('Insta-Infra UI Version: Loading version...')).toBeInTheDocument()
  })

  test('displays version when loaded successfully', async () => {
    const mockVersion = 'v2.1.0'
    GetInstaInfraVersion.mockResolvedValue(mockVersion)
    
    render(<VersionDisplay />)
    
    await waitFor(() => {
      expect(screen.getByText(`Insta-Infra UI Version: ${mockVersion}`)).toBeInTheDocument()
    })
    
    expect(GetInstaInfraVersion).toHaveBeenCalledTimes(1)
  })

  test('handles version fetch error gracefully', async () => {
    GetInstaInfraVersion.mockRejectedValue(new Error('Failed to get version'))
    
    render(<VersionDisplay />)
    
    await waitFor(() => {
      expect(screen.getByText('Insta-Infra UI Version: Error fetching version')).toBeInTheDocument()
    })
    
    // Should not show loading
    expect(screen.queryByText('Insta-Infra UI Version: Loading version...')).not.toBeInTheDocument()
  })

  test('has correct inline styles', async () => {
    const mockVersion = 'v2.1.0'
    GetInstaInfraVersion.mockResolvedValue(mockVersion)
    
    render(<VersionDisplay />)
    
    await waitFor(() => {
      const versionElement = screen.getByText(`Insta-Infra UI Version: ${mockVersion}`)
      expect(versionElement).toHaveStyle({
        backgroundColor: 'rgb(35, 39, 46)',
        color: 'rgb(169, 169, 169)', // darkgrey converts to this RGB value
        fontSize: '0.8em',
        textAlign: 'right'
      })
    })
  })
}) 