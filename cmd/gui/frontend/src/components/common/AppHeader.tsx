import React, { ChangeEvent } from 'react';

interface Props {
    header: string;
}

interface IState {
    selected: string;
}

class AppHeader extends React.Component<Props, IState>{
    constructor(props: Props) {
        super(props);
        this.state = {
            selected: "00",
        }
        this.handleSelect = this.handleSelect.bind(this);
    }


    handleSelect(e: ChangeEvent<HTMLSelectElement>) {
        alert(e.target.value)
        this.setState({
            selected: e.target.value
        })
    }

    render() {
        const { header } = this.props;
        return (
            <div className="header abs-center">
                <span>{header}</span>
                <div className="wallet-select">
                    <select onChange={this.handleSelect}>
                        <option value="01">Wallet 1</option>
                        <option value="02">Wallet 2</option>
                        <optgroup label="------------">
                            <option value="add">Add Wallet</option>
                        </optgroup>
                    </select>
                </div>
            </div>
        );
    }
}

export default AppHeader;