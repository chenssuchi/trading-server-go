// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "AdminRoleUpgrade.sol"

contract Trade is Initializable {
    uint256 private _orderId = 1;
    address public tokenAddress;
    address private _owner;

    struct Order {
        uint256 id;
        address seller;
        address buyer;
        uint256 price;
        uint256 quantity;
        bool isDael;
        uint256 saleTime;
        uint256 buyTime;
    }

    mapping(uint256 => Order) public orderInfo;

    event Buy(
        uint256 indexed orderId,
        address seller,
        address buyer,
        uint256 price,
        uint256 quantity,
        uint256 saleTime,
        uint256 buyTime
    );

    event Sale(
        uint256 indexed orderId,
        address seller,
        uint256 price,
        uint256 quantity,
        uint256 saleTime
    );

    modifier onlyOwner() {
        require(_owner == msg.sender, "Ownable: caller is not the owner");
        _;
    }

    function initialize() public initializer {
        _orderId = 1;
        _owner = msg.sender;
//        _addAdmin(msg.sender);
    }


    function adminSetOwner(address owner_) external onlyOwner
    {
        _owner = owner_;
    }

    function adminSetTokenAddress(address tokenAddress_) external onlyOwner {
        tokenAddress = tokenAddress_;
    }

    function sale(uint256 quantity, uint256 price) external {
        uint256 userBalance = IERC20(tokenAddress).balanceOf(msg.sender);
        require(userBalance >= quantity, "Insufficient balance");
        IERC20(tokenAddress).transferFrom(msg.sender, address(this), quantity);

        orderInfo[_orderId] = Order(
            _orderId,
            msg.sender,
            address(0),
            price,
            quantity,
            false,
            block.timestamp,
            0
        );

        emit Sale(
            _orderId,
            msg.sender,
            price,
            quantity,
            block.timestamp
        );

        _orderId++;
    }

    function buy(uint256 orderId_) external {
        Order storage order = orderInfo[orderId_];
        require(!order.isDael, "order status error"); // already done
        require(order.saleTime != 0, "order id error"); // orderId error

        uint256 quantity = order.quantity;
        uint256 saleTime = order.saleTime;
        uint256 price = order.price;
        address seller = order.seller;
        require(IERC20(tokenAddress).transfer(msg.sender, quantity), "server error"); // insufficient balance

        order.buyer = msg.sender;
        order.buyTime = block.timestamp;
        order.isDael = true;

        emit Buy(
            orderId_,
            seller,
            msg.sender,
            price,
            quantity,
            saleTime,
            block.timestamp
        );
    }

    function getOrderInfoById(uint256 orderId_) external view returns (Order memory){
        return orderInfo[orderId_];
    }
}