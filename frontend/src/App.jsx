import { useState, useEffect } from 'react';
import './App.css';
import { GetFilterLists, SetAlert, CancelAlert } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime';

const allMaps = [
    'Azagor', 'Basra', 'Construction', 'District', 'Dusty Dew',
    'Eduardovo', 'Frugis', 'Isle', 'Kodiak', 'Lonovo', 'Multu Islands',
    'Namak', 'Oil Dunes', 'Outskirts', 'River', 'Salhan', 'Sandy Sunset',
    'Tensa Town', 'Valley', 'Wakistan', 'Wine Paradise', 'Zalfibay'
].sort();

function CheckboxList({ title, items, selectedItems, onChange, disabled }) {
    if (items.length === 0) {
        return (
            <div className="form-group">
                <label>{title}:</label>
                <div className="list-container">
                    <span>Loading...</span>
                </div>
            </div>
        );
    }

    return (
        <div className="form-group">
            <label>{title}:</label>
            <div className="list-container">
                {items.map((item) => (
                    <div key={item} className="checkbox-item">
                        <input
                            type="checkbox"
                            id={`${title}-${item}`}
                            value={item}
                            checked={selectedItems.includes(item)}
                            onChange={onChange}
                            disabled={disabled}
                        />
                        <label htmlFor={`${title}-${item}`}>{item}</label>
                    </div>
                ))}
            </div>
        </div>
    );
}


function App() {
    const [gamemodeList, setGamemodeList] = useState([]);
    const [regionList, setRegionList] = useState([]);

    const [selectedMaps, setSelectedMaps] = useState([]);
    const [selectedGamemodes, setSelectedGamemodes] = useState([]);
    const [selectedRegions, setSelectedRegions] = useState([]);
    const [minPlayers, setMinPlayers] = useState(10);

    const [status, setStatus] = useState('No alert set.');
    const [isAlertSet, setIsAlertSet] = useState(false);
    const [statusClass, setStatusClass] = useState('');

    const resetUI = (message) => {
        setStatus(message || 'No alert set.');
        setStatusClass('');
        setIsAlertSet(false);
        setSelectedMaps([]);
        setSelectedGamemodes([]);
        setSelectedRegions([]);
        setMinPlayers(10);
    };

    useEffect(() => {
        GetFilterLists().then(lists => {
            setGamemodeList(lists.gamemodes || []);
            setRegionList(lists.regions || []);
        });

        const alertSetCleanup = EventsOn('alertSet', () => {
            setStatus(
                `Alert set for ${selectedMaps.length} maps.`
            );
            setStatusClass('status-success');
            setIsAlertSet(true);
        });

        const alertCancelledCleanup = EventsOn('alertCancelled', () => {
            resetUI('Alert cancelled by user.');
        });

        return () => {
            alertSetCleanup();
            alertCancelledCleanup();
        };
    }, [selectedMaps.length]);

    const handleSetAlert = () => {
        if (selectedMaps.length === 0) {
            setStatus('Error: Please select at least one map.');
            setStatusClass('status-error');
            return;
        }

        if (minPlayers < 1) {
            setStatus('Error: Please enter a valid player count.');
            setStatusClass('status-error');
            return;
        }

        const config = {
            maps: selectedMaps,
            gamemodes: selectedGamemodes,
            regions: selectedRegions,
            minPlayers: minPlayers
        };

        SetAlert(config);
    };

    const handleCancelAlert = () => {
        CancelAlert();
    };

    const createCheckboxHandler = (setter) => (e) => {
        const { value, checked } = e.target;
        if (checked) {
            setter((prev) => [...prev, value]);
        } else {
            setter((prev) => prev.filter((item) => item !== value));
        }
    };

    const handleMapCheckboxChange = createCheckboxHandler(setSelectedMaps);
    const handleGamemodeCheckboxChange = createCheckboxHandler(setSelectedGamemodes);
    const handleRegionCheckboxChange = createCheckboxHandler(setSelectedRegions);

    return (
        <div className="container">
            <h2>BattleBit Alerter</h2>

            <CheckboxList
                title="Select Map(s)"
                items={allMaps}
                selectedItems={selectedMaps}
                onChange={handleMapCheckboxChange}
                disabled={isAlertSet}
            />

            <CheckboxList
                title="Select Region(s) (Optional)"
                items={regionList}
                selectedItems={selectedRegions}
                onChange={handleRegionCheckboxChange}
                disabled={isAlertSet}
            />

            <CheckboxList
                title="Select Gamemode(s) (Optional)"
                items={gamemodeList}
                selectedItems={selectedGamemodes}
                onChange={handleGamemodeCheckboxChange}
                disabled={isAlertSet}
            />

            <div className="form-group">
                <label htmlFor="player-count">Min Players on Server:</label>
                <input
                    id="player-count"
                    type="number"
                    min="1"
                    value={minPlayers}
                    onChange={(e) =>
                        setMinPlayers(parseInt(e.target.value, 10) || 1)
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