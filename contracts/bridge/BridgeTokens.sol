// Copyright 2019 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.5.6;

import "../externals/openzeppelin-solidity/contracts/ownership/Ownable.sol";

contract BridgeTokens is Ownable {
    mapping(address => address) public allowedTokens; // <token, counterpart token>
    address[] public allowedTokenList;
    mapping(address => bool) public lockedTokens;

    event TokenRegistered(address indexed token);
    event TokenDeregistered(address indexed token);
    event TokenLocked(address indexed token);
    event TokenUnlocked(address indexed token);

    function getAllowedTokenList() external view returns(address[] memory) {
        return allowedTokenList;
    }

    // registerToken can update the allowed token with the counterpart token.
    function registerToken(address _token, address _cToken)
        external
        onlyOwner
    {
        require(allowedTokens[_token] == address(0));
        allowedTokens[_token] = _cToken;
        allowedTokenList.push(_token);

        emit TokenRegistered(_token);
    }

    // deregisterToken can remove the token in allowedToken list.
    function deregisterToken(address _token)
        external
        onlyOwner
    {
        require(allowedTokens[_token] != address(0));
        delete allowedTokens[_token];
        delete lockedTokens[_token];

        for (uint i = 0; i < allowedTokenList.length; i++) {
            if (allowedTokenList[i] == _token) {
                allowedTokenList[i] = allowedTokenList[allowedTokenList.length-1];
                allowedTokenList.length--;
                break;
            }
        }

        emit TokenDeregistered(_token);
    }

    // lockToken can lock the token to prevent request token transferring.
    function lockToken(address _token)
        external
        onlyOwner
    {
        require(allowedTokens[_token] != address(0));
        require(lockedTokens[_token] == false);

        lockedTokens[_token] = true;

        emit TokenLocked(_token);
    }

    // unlockToken can unlock the token to  request token transferring.
    function unlockToken(address _token)
        external
        onlyOwner
    {
        require(allowedTokens[_token] != address(0));
        require(lockedTokens[_token] == true);

        lockedTokens[_token] = false;

        emit TokenUnlocked(_token);
    }
}
