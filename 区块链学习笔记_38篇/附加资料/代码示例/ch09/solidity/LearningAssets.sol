// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC20Permit} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";

// Teaching assets for a local chain. No admin minting or upgrade mechanism.
contract LearningToken is ERC20, ERC20Permit {
    constructor() ERC20("Learning Token", "LRN") ERC20Permit("Learning Token") {
        _mint(msg.sender, 1000 * 10 ** decimals());
    }
}

contract LearningNFT is ERC721 {
    constructor() ERC721("Learning NFT", "LNFT") {
        _safeMint(msg.sender, 1);
        _safeMint(msg.sender, 2);
    }
}
