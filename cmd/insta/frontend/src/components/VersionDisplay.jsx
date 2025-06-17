import { useState, useEffect } from 'react';
import { getApiInfo } from '../api/client';

function VersionDisplay() {
    const [version, setVersion] = useState('Loading version...');

    useEffect(() => {
        getApiInfo()
            .then(info => setVersion(info.version || 'Unknown'))
            .catch(err => {
                console.error("Error fetching version:", err);
                setVersion('Error fetching version');
            });
    }, []);

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