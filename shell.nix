{ pkgs ? import <nixpkgs> {} }:


  pkgs.mkShell {
    buildInputs = with pkgs; [

      go_1_23
      python3
      python3Packages.pip
      
      sqlite
      sqlc
      
      docker
      docker-compose
      
      protobuf
      protoc-gen-go
      protoc-gen-go-grpc
      
      git
      jq
    ];

  shellHook = ''
    # Create a local bin directory and symlink the protoc plugins
    mkdir -p $HOME/.local/bin
    
    # Create symlinks with the exact names protoc expects
    ln -sf ${pkgs.protoc-gen-go}/bin/protoc-gen-go $HOME/.local/bin/protoc-gen-go
    ln -sf ${pkgs.protoc-gen-go-grpc}/bin/protoc-gen-go-grpc $HOME/.local/bin/protoc-gen-go-grpc
    
    # Add to PATH
    export PATH="$HOME/.local/bin:$PATH"
    
    echo "🔧 Development environment loaded"
    echo "Go version: $(go version)"
    echo "Protoc version: $(protoc --version)"
    
    # Verify plugins are available
    which protoc-gen-go
    which protoc-gen-go-grpc
    
    # Create alias for convenience
    alias proto-gen='protoc --go_out=. --go_grpc_out=. proto/*.proto'
  '';
  }
