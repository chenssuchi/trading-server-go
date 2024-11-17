// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract IntoToken is ERC20 {
    constructor(uint256 initialSupply) ERC20("Into Token", "IT") {
        _mint(msg.sender, initialSupply);
    }
}