#!/bin/bash
# ============================================
# AxionPay WARP SETUP (Client Machine)
# Installs and Configures Cloudflare WARP for Zero Trust Access
# ============================================

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=== AxionPay WARP Client Setup ===${NC}"

# 1. Install WARP
echo -e "\n${YELLOW}[1/3] Installing Cloudflare WARP...${NC}"

# Add GPG Key
curl -fsSL https://pkg.cloudflareclient.com/pubkey.gpg | sudo gpg --yes --dearmor --output /usr/share/keyrings/cloudflare-warp-archive-keyring.gpg

# Add Repo
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/cloudflare-warp-archive-keyring.gpg] https://pkg.cloudflareclient.com/ $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/cloudflare-client.list

# Install
sudo apt-get update && sudo apt-get install cloudflare-warp -y
echo -e "${GREEN}✓ WARP Installed${NC}"

# 2. Register
echo -e "\n${YELLOW}[2/3] Registering Device...${NC}"
echo -e "${YELLOW}!!! ACTION REQUIRED !!!${NC}"
echo -e "A registration URL will appear below."
echo -e "Copy it, open in your browser, and log in with Cloudflare Zero Trust."
echo -e "Press ENTER after you have authorized the device in the browser."
echo ""

warp-cli registration new

read -p "Press ENTER after successful authentication..."

# 3. Connect
echo -e "\n${YELLOW}[3/3] Connecting to Zero Trust Network...${NC}"
warp-cli connect
warp-cli mode warp

# Wait for connection
sleep 3
STATUS=$(warp-cli status | grep Status)
echo -e "${GREEN}Current Status: $STATUS${NC}"

echo -e "\n${GREEN}=== SETUP COMPLETE ===${NC}"
echo -e "Try connecting now using the private IP:"
echo -e "${YELLOW}./bin/pixelcraft 10.0.0.5 <HYDRA_SECRET>${NC}"
