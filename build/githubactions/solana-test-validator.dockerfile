FROM ubuntu:20.04

# Update default packages
RUN apt-get update

# Install Ubuntu packages
RUN apt-get install -y \
    build-essential \
    curl \ 
    wget

# Install Rust
RUN curl https://sh.rustup.rs -sSf | bash -s -- -y
ENV PATH /root/.cargo/bin:$PATH
RUN rustup component add rustfmt

# Install Solana CLI
RUN wget -q https://github.com/solana-labs/solana/releases/download/v1.8.4/solana-release-x86_64-unknown-linux-gnu.tar.bz2
RUN tar jxf solana-release-x86_64-unknown-linux-gnu.tar.bz2
ENV PATH /solana-release/bin:$PATH
RUN yes | solana-keygen new
ENV KEEPER_BOT_ACCOUNT $(cat /root/.config/solana/id.json)