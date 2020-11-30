import React from 'react';

interface Props {

}

const AppFooter: React.FC<Props> = ({ }) => {
    return (
        <footer>
            <span>Olympus vX.XX</span>
            <span className="fas-icon">signal</span>
        </footer>
    );
}

export default AppFooter;