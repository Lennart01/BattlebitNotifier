import { useState, useEffect } from 'react';
import './App.css';
import { GetMapList, SetAlert, CancelAlert } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime';

function App() {
    const [mapList, setMapList] = useState([]);
    const [selectedMaps, setSelectedMaps] = useState([]);
    const [playerCount, setPlayerCount] = useState(100);
    const [status, setStatus] = useState('No alert set.');
    const [isAlertSet, setIsAlertSet] = useState(false);
    const [statusClass, setStatusClass] = useState('');

    const resetUI = (message) => {
        setStatus(message || 'No alert set.');
        setStatusClass('');
        setIsAlertSet(false);
        setSelectedMaps([]);
        setPlayerCount(100);
    };

    useEffect(() => {
        GetMapList().then((maps) => setMapList(maps || []));

        const mapsUpdatedCleanup = EventsOn('mapsUpdated', (maps) => {
            setMapList(maps || []);
        });

        const alertTriggeredCleanup = EventsOn('alertTriggered', () => {
            resetUI('Alert triggered! Ready for new alert.');
        });

        return () => {
            mapsUpdatedCleanup();
            alertTriggeredCleanup();
        };
    }, []);

    const handleSetAlert = () => {
        if (selectedMaps.length === 0) {
            setStatus('Error: Please select at least one map.');
            setStatusClass('status-error');
            return;
        }

        if (playerCount < 1) {
            setStatus('Error: Please enter a valid player count.');
            setStatusClass('status-error');
            return;
        }

        SetAlert(selectedMaps, playerCount);
        setStatus(
            `Alert set for ${selectedMaps.length} maps @ ${playerCount} players.`
        );
        setStatusClass('status-success');
        setIsAlertSet(true);
    };

    const handleCancelAlert = () => {
        CancelAlert();
        resetUI('Alert cancelled by user.');
    };

    const handleMapCheckboxChange = (e) => {
        const { value, checked } = e.target;
        if (checked) {
            setSelectedMaps((prev) => [...prev, value]);
        } else {
            setSelectedMaps((prev) => prev.filter((map) => map !== value));
        }
    };

    return (
        <div className="container">
            <h2>BattleBit Alerter</h2>
            <div className="form-group">
                <label>Select Map(s):</label>
                <div id="map-list-container">
                    {mapList.length === 0 && (
                        <span>Loading maps...</span>
                    )}
                    {mapList.map((mapName) => (
                        <div key={mapName} className="checkbox-item">
                            <input
                                type="checkbox"
                                id={`map-${mapName}`}
                                value={mapName}
                                checked={selectedMaps.includes(mapName)}
                                onChange={handleMapCheckboxChange}
                                disabled={isAlertSet}
                            />
                            <label htmlFor={`map-${mapName}`}>{mapName}</label>
                        </div>
                    ))}
                </div>
            </div>

            <div className="form-group">
                <label htmlFor="player-count">Min Total Players:</label>
                <input
                    id="player-count"
                    type="number"
                    min="1"
                    value={playerCount}
                    onChange={(e) =>
                        setPlayerCount(parseInt(e.target.value, 10) || 1)
                    }
                    disabled={isAlertSet}
                />
            </div>

            {!isAlertSet ? (
                <button id="set-alert-btn" onClick={handleSetAlert}>
                    Set Alert
                </button>
            ) : (
                <button
                    id="cancel-alert-btn"
                    className="btn-danger"
                    onClick={handleCancelAlert}>
                    Cancel Alert
                </button>
            )}

            <div id="status" className={statusClass}>
                {status}
            </div>
        </div>
    );
}

export default App;