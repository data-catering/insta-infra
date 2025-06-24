import { useState, useEffect } from 'react';
import { getApiInfo } from '../api/client';
import { useErrorHandler } from './ErrorMessage';

function VersionDisplay() {
    const [version, setVersion] = useState('Loading version...');
    const { addError } = useErrorHandler();

    useEffect(() => {
        getApiInfo()
            .then(info => setVersion(info.version || 'Unknown'))
            .catch(err => {
                console.error("Error fetching version:", err);
                setVersion('Error fetching version');
                
                // Add detailed error message for user
                addError({
                    type: 'warning',
                    title: 'Version Information Unavailable',
                    message: 'Unable to fetch application version information',
                    details: `Error: ${err.message || err.toString()} | Time: ${new Date().toLocaleTimeString()}`,
                    metadata: {
                        action: 'fetch_version',
                        errorMessage: err.message || err.toString(),
                        timestamp: new Date().toISOString()
                    },
                    actions: [
                        {
                            label: 'Retry',
                            onClick: () => {
                                setVersion('Loading version...');
                                getApiInfo()
                                    .then(info => setVersion(info.version || 'Unknown'))
                                    .catch(() => setVersion('Error fetching version'));
                            },
                            variant: 'secondary'
                        }
                    ]
                });
            });
    }, [addError]);

    return (
        <div 
            style={{
                width: '100%',
                padding: '5px 10px',
                backgroundColor: '#23272e', // Slightly different from main background for subtle separation
                color: 'darkgrey',
                fontSize: '0.8em',
                textAlign: 'right',
                borderBottom: '1px solid #444'
            }}
        >
            Insta-Infra UI Version: {version}
        </div>
    );
}

export default VersionDisplay; 